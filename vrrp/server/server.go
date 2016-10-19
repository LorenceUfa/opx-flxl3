//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package server

import (
	"asicd/asicdCommonDefs"
	"asicdServices"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/google/gopacket"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"utils/ipcutils"
	"utils/logging"
	"vrrpd"
)

/*
func (svr *VrrpServer) VrrpUpdateIntfIpAddr(gblInfo *VrrpGlobalInfo) bool {
	IpAddr, ok := svr.vrrpIfIndexIpAddr[gblInfo.IntfConfig.IfIndex]
	if ok == false {
		svr.logger.Err(fmt.Sprintln("missed ipv4 intf notification for IfIndex:",
			gblInfo.IntfConfig.IfIndex))
		gblInfo.IpAddr = ""
		return false
	}
	gblInfo.IpAddr = IpAddr
	return true
}
*/

func (svr *VrrpServer) VrrpPopulateIntfState(key string, entry *vrrpd.VrrpIntfState) bool {
	gblInfo, ok := svr.vrrpGblInfo[key]
	if ok == false {
		svr.logger.Err(fmt.Sprintln("Entry not found for", key))
		return ok
	}
	entry.IfIndex = gblInfo.IntfConfig.IfIndex
	entry.VRID = gblInfo.IntfConfig.VRID
	entry.IntfIpAddr = gblInfo.IpAddr
	entry.Priority = gblInfo.IntfConfig.Priority
	entry.VirtualIPv4Addr = gblInfo.IntfConfig.VirtualIPv4Addr
	entry.AdvertisementInterval = gblInfo.IntfConfig.AdvertisementInterval
	entry.PreemptMode = gblInfo.IntfConfig.PreemptMode
	entry.VirtualRouterMACAddress = gblInfo.VirtualRouterMACAddress
	entry.SkewTime = gblInfo.SkewTime
	entry.MasterDownTimer = gblInfo.MasterDownValue
	gblInfo.StateNameLock.Lock()
	entry.VrrpState = gblInfo.StateName
	gblInfo.StateNameLock.Unlock()
	return ok
}

func (svr *VrrpServer) VrrpPopulateVridState(key string, entry *vrrpd.VrrpVridState) bool {
	gblInfo, ok := svr.vrrpGblInfo[key]
	if ok == false {
		svr.logger.Err(fmt.Sprintln("Entry not found for", key))
		return ok
	}
	entry.IfIndex = gblInfo.IntfConfig.IfIndex
	entry.VRID = gblInfo.IntfConfig.VRID
	gblInfo.StateInfoLock.Lock()
	entry.AdverRx = int32(gblInfo.StateInfo.AdverRx)
	entry.AdverTx = int32(gblInfo.StateInfo.AdverTx)
	entry.CurrentState = gblInfo.StateInfo.CurrentFsmState
	entry.PreviousState = gblInfo.StateInfo.PreviousFsmState
	entry.LastAdverRx = gblInfo.StateInfo.LastAdverRx
	entry.LastAdverTx = gblInfo.StateInfo.LastAdverTx
	entry.MasterIp = gblInfo.StateInfo.MasterIp
	entry.TransitionReason = gblInfo.StateInfo.ReasonForTransition
	gblInfo.StateInfoLock.Unlock()
	return ok
}

func (svr *VrrpServer) VrrpDeleteGblInfo(config vrrpd.VrrpIntf) {
	key := strconv.Itoa(int(config.IfIndex)) + "_" + strconv.Itoa(int(config.VRID))
	gblInfo, found := svr.vrrpGblInfo[key]
	if found {
		svr.VrrpUpdateSubIntf(gblInfo, false /*disable*/)
	}
	delete(svr.vrrpGblInfo, key)
	for i := 0; i < len(svr.vrrpIntfStateSlice); i++ {
		if svr.vrrpIntfStateSlice[i] == key {
			svr.vrrpIntfStateSlice = append(svr.vrrpIntfStateSlice[:i],
				svr.vrrpIntfStateSlice[i+1:]...)
			break
		}
	}
	if len(svr.vrrpIntfStateSlice) != 0 {
		return
	}
	svr.logger.Info("No more vrrp configured, disabling protocol mac")
	svr.VrrpUpdateProtocolMacEntry(false /*delete vrrp protocol mac*/)
}

func (svr *VrrpServer) VrrpUpdateIntf(origconfig vrrpd.VrrpIntf,
	newconfig vrrpd.VrrpIntf, attrset []bool) {
	key := strconv.Itoa(int(origconfig.IfIndex)) + "_" +
		strconv.Itoa(int(origconfig.VRID))
	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No object for " + key)
		return
	}
	/*
		0	1 : i32 IfIndex
		1	2 : i32 VRID
		2	3 : i32 Priority
		3	4 : string VirtualIPv4Addr
		4	5 : i32 AdvertisementInterval
		5	6 : bool PreemptMode
		6	7 : bool AcceptMode
	*/
	updDownTimer := false
	for elem, _ := range attrset {
		//for elem <= VRRP_TOTAL_INTF_CONFIG_ELEMENTS {
		if !attrset[elem] {
			continue
		} else {
			switch elem {
			case 0:
				// Cannot change IfIndex
			case 1:
				// Cannot change VRID
			case 2:
				gblInfo.IntfConfig.Priority = newconfig.Priority
			case 3:
				gblInfo.IntfConfig.VirtualIPv4Addr =
					newconfig.VirtualIPv4Addr
			case 4:
				gblInfo.IntfConfig.AdvertisementInterval =
					newconfig.AdvertisementInterval
				updDownTimer = true
			case 5:
				gblInfo.IntfConfig.PreemptMode = newconfig.PreemptMode
			case 6:
				gblInfo.IntfConfig.AcceptMode = newconfig.AcceptMode
			}
		}
	}

	// If Advertisment value changed then we need to update master down timer
	if updDownTimer {
		gblInfo.MasterDownLock.Lock()
		svr.VrrpCalculateDownValue(gblInfo.IntfConfig.AdvertisementInterval,
			&gblInfo)
		gblInfo.MasterDownLock.Unlock()
		svr.vrrpGblInfo[key] = gblInfo
		svr.VrrpHandleMasterDownTimer(key)
	} else {
		svr.vrrpGblInfo[key] = gblInfo
	}
}

func (svr *VrrpServer) VrrpGetBulkVrrpIntfStates(idx int, cnt int) (int, int, []vrrpd.VrrpIntfState) {
	var nextIdx int
	var count int
	if svr.vrrpIntfStateSlice == nil {
		svr.logger.Info("Interface Slice is not initialized")
		return 0, 0, nil
	}
	length := len(svr.vrrpIntfStateSlice)
	result := make([]vrrpd.VrrpIntfState, cnt)
	var i int
	var j int

	for i, j = 0, idx; i < cnt && j < length; j++ {
		key := svr.vrrpIntfStateSlice[j]
		_ = svr.VrrpPopulateIntfState(key, &result[i])
		i++
	}
	if j == length {
		nextIdx = 0
	}
	count = i
	return nextIdx, count, result
}

func (svr *VrrpServer) VrrpGetBulkVrrpVridStates(idx int, cnt int) (int, int, []vrrpd.VrrpVridState) {
	var nextIdx int
	var count int
	if svr.vrrpIntfStateSlice == nil {
		svr.logger.Info("Interface slice is not initialized")
		return 0, 0, nil
	}
	length := len(svr.vrrpIntfStateSlice)
	result := make([]vrrpd.VrrpVridState, cnt)
	var i int
	var j int

	for i, j = 0, idx; i < cnt && j < length; j++ {
		key := svr.vrrpIntfStateSlice[j]
		_ = svr.VrrpPopulateVridState(key, &result[i])
		i++
	}
	if j == length {
		nextIdx = 0
	}
	count = i
	return nextIdx, count, result
}

func (svr *VrrpServer) VrrpMapIfIndexToLinuxIfIndex(IfIndex int32) {
	_, found := svr.vrrpLinuxIfIndex2AsicdIfIndex[IfIndex]
	if found {
		return
	}
	vlanId := asicdCommonDefs.GetIntfIdFromIfIndex(IfIndex)
	vlanName, ok := svr.vrrpVlanId2Name[vlanId]
	if ok == false {
		svr.logger.Err(fmt.Sprintln("no mapping for vlan", vlanId))
		return
	}
	linuxInterface, err := net.InterfaceByName(vlanName)
	if err != nil {
		svr.logger.Err(fmt.Sprintln("Getting linux If index for",
			"IfIndex:", IfIndex, "failed with ERROR:", err))
		return
	}
	svr.logger.Info(fmt.Sprintln("Linux Id:", linuxInterface.Index,
		"maps to IfIndex:", IfIndex))
	svr.vrrpLinuxIfIndex2AsicdIfIndex[IfIndex] = linuxInterface
}

func (svr *VrrpServer) VrrpCloseAllPcapHandlers() {
	for i := 0; i < len(svr.vrrpIntfStateSlice); i++ {
		key := svr.vrrpIntfStateSlice[i]
		gblInfo := svr.vrrpGblInfo[key]
		if gblInfo.pHandle != nil {
			gblInfo.pHandle.Close()
		}
	}
}

/*
func (svr *VrrpServer) VrrpConnectToAsicd(client VrrpClientJson) error {
	svr.logger.Info(fmt.Sprintln("VRRP: Connecting to asicd at port",
		client.Port))
	var err error
	svr.asicdClient.Address = "localhost:" + strconv.Itoa(client.Port)
	svr.asicdClient.Transport, svr.asicdClient.PtrProtocolFactory, err =
		ipcutils.CreateIPCHandles(svr.asicdClient.Address)
	if svr.asicdClient.Transport == nil ||
		svr.asicdClient.PtrProtocolFactory == nil ||
		err != nil {
		return err
	}
	svr.asicdClient.ClientHdl =
		asicdServices.NewASICDServicesClientFactory(
			svr.asicdClient.Transport,
			svr.asicdClient.PtrProtocolFactory)
	svr.asicdClient.IsConnected = true
	return nil
}

func (svr *VrrpServer) VrrpConnectToUnConnectedClient(client VrrpClientJson) error {
	switch client.Name {
	case "asicd":
		return svr.VrrpConnectToAsicd(client)
	default:
		return errors.New(VRRP_CLIENT_CONNECTION_NOT_REQUIRED)
	}
}
func (svr *VrrpServer) VrrpConnectAndInitPortVlan() error {
	configFile := svr.paramsDir + "/clients.json"
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		svr.logger.Err(fmt.Sprintln("VRRP:Error while reading configuration file",
			configFile))
		return err
	}
	var unConnectedClients []VrrpClientJson
	err = json.Unmarshal(bytes, &unConnectedClients)
	if err != nil {
		svr.logger.Err("VRRP: Error in Unmarshalling Json")
		return err
	}
	re_connect := 25
	count := 0
	// connect to client
	for {
		time.Sleep(time.Millisecond * 500)
		for i := 0; i < len(unConnectedClients); i++ {
			err := svr.VrrpConnectToUnConnectedClient(
				unConnectedClients[i])
			if err == nil {
				svr.logger.Info("VRRP: Connected to " +
					unConnectedClients[i].Name)
				unConnectedClients = append(
					unConnectedClients[:i],
					unConnectedClients[i+1:]...)

			} else if err.Error() ==
				VRRP_CLIENT_CONNECTION_NOT_REQUIRED {
				unConnectedClients = append(
					unConnectedClients[:i],
					unConnectedClients[i+1:]...)
			} else {
				count++
				if count == re_connect {
					svr.logger.Err(fmt.Sprintln(
						"Connecting to",
						unConnectedClients[i].Name,
						"failed ", err))
					count = 0
				}
			}
		}
		if len(unConnectedClients) == 0 {
			break
		}
	}

	svr.VrrpGetInfoFromAsicd()

	// OS Signal channel listener thread
	svr.VrrpOSSignalHandle()
	return err
}
*/

/*
func (svr *VrrpServer) VrrpValidateIntfConfig(IfIndex int32) error {
	// Check Vlan is created
	vlanId := asicdCommonDefs.GetIntfIdFromIfIndex(IfIndex)
	_, created := svr.vrrpVlanId2Name[vlanId]
	if !created {
		return errors.New(VRRP_VLAN_NOT_CREATED)
	}

	// Check ipv4 interface is created
	_, created = svr.vrrpLinuxIfIndex2AsicdIfIndex[IfIndex]
	if !created {
		return errors.New(VRRP_IPV4_INTF_NOT_CREATED)
	}

	return nil
}
*/

func (svr *VrrpServer) VrrpChecknUpdateGblInfo(IfIndex int32, IpAddr string) {
	for _, key := range svr.vrrpIntfStateSlice {
		startFsm := false
		splitString := strings.Split(key, "_")
		// splitString = { IfIndex, VRID }
		ifindex, _ := strconv.Atoi(splitString[0])
		if int32(ifindex) != IfIndex {
			// Key doesn't match
			continue
		}
		// If IfIndex matches then use that key and check if gblInfo is
		// created or not
		gblInfo, found := svr.vrrpGblInfo[key]
		if !found {
			svr.logger.Err("No entry found for Ifindex:" +
				splitString[0] + " VRID:" + splitString[1] +
				" hence not updating ip addr, " +
				"it will be updated during create")
			continue
		}
		gblInfo.IpAddr = IpAddr
		gblInfo.StateNameLock.Lock()
		if gblInfo.StateName == VRRP_UNINTIALIZE_STATE {
			startFsm = true
			gblInfo.StateName = VRRP_INITIALIZE_STATE
		}
		gblInfo.StateNameLock.Unlock()
		svr.vrrpGblInfo[key] = gblInfo
		// Create Pkt Listener if not created... This will handle a
		// scneario when VRRP configs are done before IF Index is up
		gblInfo.PcapHdlLock.Lock()
		if gblInfo.pHandle == nil {
			gblInfo.PcapHdlLock.Unlock()
			svr.VrrpInitPacketListener(key, IfIndex)
		} else {
			gblInfo.PcapHdlLock.Unlock()
		}
		if !svr.vrrpMacConfigAdded {
			svr.logger.Info("Adding protocol mac for punting packets to CPU")
			svr.VrrpUpdateProtocolMacEntry(true /*add vrrp protocol mac*/)
		}
		if startFsm {
			svr.vrrpFsmCh <- VrrpFsm{
				key: key,
			}
		}
	}
}

func (svr *VrrpServer) HandlerCreateConfig(cfg *config.IntfCfg) {
	key := KeyInfo{cfg.IntfRef, cfg.Version}
	intf, exists := svr.Intf[key]
	if exists {
		debug.Logger.Err("During Create we should not any entry in the DB")
		return
	}
	l3Info := &L3Intf{}
	l3Info.IfName = cfg.IntfRef
	switch cfg.Version {
	case config.VERSION2:
		ifIndex, exists := svr.V4IntfRefToIfIndex[cfg.IntfRef]
		if exists {
			l3Info.IfIndex = ifIndex
			v4, exists := svr.V4[ifIndex]
			if exists {
				l3Info.IpAddr = v4.Cfg.IpAddr
			}
		}
		// if cross reference exists then only set l3Info else just pass go defaults and it will updated
		// later once we have configured ipv4 or ipv6 interface
	case config.VERSION3:
		ifIndex, exists := svr.V6IntfRefToIfIndex[cfg.IntfRef]
		l3Info.IfIndex = ifIndex
		if exists {
			v6, exists := svr.V6[ifIndex]
			if exists {
				// @TODO: do we have to use linkscope ip or global scope ip Check RFC
				l3Info.IpAddr = v6.Cfg.IpAddr
			}
		}
	}
	intf.Init(cfg, l3Info, svr.StateCh)
}

func (svr *VrrpServer) HandleVrrpConfig(cfg *config.IntfCfg) {
	switch cfg.Operation {
	case config.CREATE:
		svr.HandlerCreateConfig(cfg)
	case config.UPDATE:

	case config.DELETE:
	}
	/*
		key := strconv.Itoa(int(config.IfIndex)) + "_" + strconv.Itoa(int(config.VRID))
		gblInfo := svr.vrrpGblInfo[key]

		gblInfo.IntfConfig.IfIndex = config.IfIndex
		gblInfo.IntfConfig.VRID = config.VRID
		gblInfo.IntfConfig.VirtualIPv4Addr = config.VirtualIPv4Addr
		gblInfo.IntfConfig.PreemptMode = config.PreemptMode
		gblInfo.IntfConfig.Priority = config.Priority
		if config.AdvertisementInterval == 0 {
			gblInfo.IntfConfig.AdvertisementInterval = 1
		} else {
			gblInfo.IntfConfig.AdvertisementInterval = config.AdvertisementInterval
		}

		if config.AcceptMode == true {
			gblInfo.IntfConfig.AcceptMode = true
		} else {
			gblInfo.IntfConfig.AcceptMode = false
		}

		if gblInfo.IntfConfig.VRID < 10 {
			gblInfo.VirtualRouterMACAddress = VRRP_IEEE_MAC_ADDR +
				"0" + strconv.Itoa(int(gblInfo.IntfConfig.VRID))

		} else {
			gblInfo.VirtualRouterMACAddress = VRRP_IEEE_MAC_ADDR +
				strconv.Itoa(int(gblInfo.IntfConfig.VRID))
		}

		// Initialize Locks for accessing shared ds
		gblInfo.PcapHdlLock = &sync.RWMutex{}
		gblInfo.StateNameLock = &sync.RWMutex{}
		gblInfo.MasterDownLock = &sync.RWMutex{}
		gblInfo.StateInfoLock = &sync.RWMutex{}

		// Update Ip Addr at last
		svr.VrrpUpdateIntfIpAddr(&gblInfo)

		// Set Initial state
		gblInfo.StateNameLock.Lock()
		gblInfo.StateName = VRRP_INITIALIZE_STATE
		gblInfo.StateNameLock.Unlock()
		svr.vrrpGblInfo[key] = gblInfo
		svr.vrrpIntfStateSlice = append(svr.vrrpIntfStateSlice, key)

		// Create Packet listener first so that pcap handler is created...
		// We will not receive any vrrp packets as punt to CPU is not yet done
		svr.VrrpInitPacketListener(key, config.IfIndex)

		// Register Protocol Mac
		if !svr.vrrpMacConfigAdded {
			svr.logger.Info("Adding protocol mac for punting packets to CPU")
			svr.VrrpUpdateProtocolMacEntry(true /*add vrrp protocol mac*/ /*)
	}
	// Start FSM
	svr.vrrpFsmCh <- VrrpFsm{
		key: key,
	}
	*/
}

func (svr *VrrpServer) UpdateProtocolMacEntry(add bool) {
	switch add {
	case true:
		svr.SwitchPlugin.EnablePacketReception(packet.VRRP_PROTOCOL_MAC)
	case false:
		svr.SwitchPlugin.DisablePacketReception(packet.VRRP_PROTOCOL_MAC)
		// @TODO: tear down all fsm states and configuration
	}
}

func (svr *VrrpServer) HandlerGlobalConfig(gCfg *config.GlobalConfig) {
	debug.Logger.Info("Handling Global Config for:", *gCfg)
	switch gCfg.Operation {
	case config.CREATE:
		debug.Logger.Info("Vrrp Enabled, configuring Protocol Mac")
		svr.UpdateProtocolMacEntry(true /*Enable*/)
	case config.UPDATE:
		debug.Logger.Info("Vrrp Disabled, deleting Protocol Mac")
		svr.UpdateProtocolMacEntry(false /*Enable*/)
	}
}

func (svr *VrrpServer) EventListener() {
	// Start receviing in rpc values in the channell
	for {
		select {

		case gCfg, ok := <-svr.GblCfgCh:
			if ok {
				svr.HandlerGlobalConfig(gCfg)
			}
		case cfg, ok := <-svr.CfgCh:
			if ok {
				svr.HandleVrrpConfig(cfg)
			}
			/*
				case rxChInfo, ok := <-svr.RxPktCh:
					if ok {
						svr.ProcessRxPkt(rxChInfo)
					}
			*/
			/*
				case intfConf := <-svr.VrrpCreateIntfConfigCh:
					svr.VrrpCreateGblInfo(intfConf)
				case delConf := <-svr.VrrpDeleteIntfConfigCh:
					svr.VrrpDeleteGblInfo(delConf)
				case fsmInfo := <-svr.vrrpFsmCh:
					svr.VrrpFsmStart(fsmInfo)
				case sendInfo := <-svr.vrrpTxPktCh:
					svr.VrrpSendPkt(sendInfo.key, sendInfo.priority)
				case rcvdInfo := <-svr.vrrpRxPktCh:
					svr.VrrpCheckRcvdPkt(rcvdInfo.pkt, rcvdInfo.key,
						rcvdInfo.IfIndex)
				case updConfg := <-svr.VrrpUpdateIntfConfigCh:
					svr.VrrpUpdateIntf(updConfg.OldConfig, updConfg.NewConfig,
						updConfg.AttrSet)
			*/
		}

	}
}

func (svr *VrrpServer) GetSystemInfo() {
	// Get ports information
	svr.GetPorts()
	// Get vlans information
	svr.GetVlans()
	// Get IP Information
	svr.GetIPIntfs()
}

func (svr *VrrpServer) InitGlobalDS() {
	svr.L2Port = make(map[int32]config.PhyPort, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V6 = make(map[int32]V6Intf, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V4 = make(map[int32]V4Intf, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	//svr.L3Port = make(map[int32]IpIntf, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.VlanInfo = make(map[int32]config.VlanInfo, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.Intf = make(map[KeyInfo]VrrpInterface, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.CfgCh = make(chan *config.IntfCfg, VRRP_INTF_CONFIG_CH_SIZE)
	svr.V4IntfRefToIfIndex = make(map[string]int32, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V6IntfRefToIfIndex = make(map[string]int32, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.StateCh = make(chan StateInfo, VRRP_RX_BUF_CHANNEL_SIZE)
	svr.GblCfgCh = make(chan config.GlobalConfig)
	/*
		svr.vrrpGblInfo = make(map[string]VrrpGlobalInfo, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
		svr.vrrpIfIndexIpAddr = make(map[int32]string, VRRP_INTF_IPADDR_MAPPING_DEFAULT_SIZE)
		svr.vrrpLinuxIfIndex2AsicdIfIndex = make(map[int32]*net.Interface, VRRP_LINUX_INTF_MAPPING_DEFAULT_SIZE)
		svr.vrrpVlanId2Name = make(map[int]string, VRRP_VLAN_MAPPING_DEFAULT_SIZE)
		svr.VrrpCreateIntfConfigCh = make(chan vrrpd.VrrpIntf, VRRP_INTF_CONFIG_CH_SIZE)
		svr.VrrpDeleteIntfConfigCh = make(chan vrrpd.VrrpIntf, VRRP_INTF_CONFIG_CH_SIZE)
		svr.vrrpRxPktCh = make(chan VrrpPktChannelInfo, VRRP_RX_BUF_CHANNEL_SIZE)
		svr.vrrpTxPktCh = make(chan VrrpTxChannelInfo, VRRP_TX_BUF_CHANNEL_SIZE)
		svr.VrrpUpdateIntfConfigCh = make(chan VrrpUpdateConfig, VRRP_INTF_CONFIG_CH_SIZE)
		svr.vrrpFsmCh = make(chan VrrpFsm, VRRP_FSM_CHANNEL_SIZE)
		svr.vrrpSnapshotLen = 1024
		svr.vrrpPromiscuous = false
		svr.vrrpTimeout = 10 * time.Microsecond
		svr.vrrpMacConfigAdded = false
		svr.vrrpPktSend = make(chan bool)
	*/
}

func (svr *VrrpServer) DeAllocateMemory() {
	svr.vrrpGblInfo = nil
	svr.vrrpIfIndexIpAddr = nil
	svr.vrrpLinuxIfIndex2AsicdIfIndex = nil
	svr.vrrpVlanId2Name = nil
	svr.vrrpRxPktCh = nil
	svr.vrrpTxPktCh = nil
	svr.VrrpDeleteIntfConfigCh = nil
	svr.VrrpCreateIntfConfigCh = nil
	svr.VrrpUpdateIntfConfigCh = nil
	svr.vrrpFsmCh = nil
}

func (svr *VrrpServer) signalHandler(sigChannel <-chan os.Signal) {
	signal := <-sigChannel
	switch signal {
	case syscall.SIGHUP:
		debug.Logger.Alert("Received SIGHUP Signal")
		//svr.VrrpCloseAllPcapHandlers()
		svr.DeAllocateMemory()
		debug.Logger.Alert("Closed all pcap's and freed memory")
		os.Exit(0)
	default:
		debug.Logger.Info("Unhandled Signal:", signal)
	}
}

func (svr *VrrpServer) OSSignalHandle() {
	sigChannel := make(chan os.Signal, 1)
	signalList := []os.Signal{syscall.SIGHUP}
	signal.Notify(sigChannel, signalList...)
	go svr.signalHandler(sigChannel)
}

/*

*/

func (svr *VrrpServer) VrrpStartServer(paramsDir string) {
	svr.OSSignalHandle()
	svr.ReadDB()
	svr.InitGlobalDS()
	svr.GetSystemInfo()
	go svr.EventListener()
	/*
		svr.paramsDir = paramsDir
		// First connect to client to avoid any issues with start/re-start
		svr.VrrpConnectAndInitPortVlan()

		// Initialize DB
		err := svr.VrrpInitDB()
		if err != nil {
			svr.logger.Err("VRRP: DB init failed")
		} else {
			// Populate Gbl Configs
			svr.VrrpReadDB()
			svr.VrrpCloseDB()
		}
		go svr.VrrpChannelHanlder()
	*/
}

func VrrpNewServer(sPlugin asicdClient.AsicdClientIntf, dmnBase *dmnBase.FSBaseDmn) *VrrpServer {
	svr := &VrrpServer{}
	svr.SwitchPlugin = sPlugin
	svr.dmnBase = dmnBase
	// Allocate memory to all the Data Structures
	//vrrpServerInfo.VrrpInitGlobalDS()
	return svr
}
