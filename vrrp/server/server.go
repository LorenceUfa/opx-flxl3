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
	"encoding/json"
	"errors"
	"fmt"
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
*/

/*
func (svr *VrrpServer) VrrpDeleteGblInfo(config vrrpd.VrrpIntf) {
	key := strconv.Itoa(int(config.IfIndex)) + "_" + strconv.Itoa(int(config.VRID))
	gblInfo, found := svr.vrrpGblInfo[key]
	if found {
		svr.VrrpUpdateSubIntf(gblInfo, false /*disable*/ //)
/*
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
	svr.VrrpUpdateProtocolMacEntry(false /*delete vrrp protocol mac*/ //)
/*
}
*/
/*
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
/*
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
*/

func (svr *VrrpServer) EventListener() {
	// Start receviing in rpc values in the channell
	for {
		select {

		case gCfg, ok := <-svr.GblCfgCh:
			if ok {
				svr.HandleGlobalConfig(gCfg)
			}
		case cfg, ok := <-svr.CfgCh:
			if ok {
				svr.HandleIntfConfig(cfg)
			}
		case stInfo, ok := <-svr.StateCh:
			if ok {
				svr.HandleStateUpdate(stInfo)
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
	svr.VlanInfo = make(map[int32]config.VlanInfo, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.Intf = make(map[KeyInfo]VrrpInterface, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V4IntfRefToIfIndex = make(map[string]int32, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V6IntfRefToIfIndex = make(map[string]int32, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.StateCh = make(chan *IntfState, VRRP_RX_BUF_CHANNEL_SIZE)
	svr.GblCfgCh = make(chan config.GlobalConfig)
	svr.CfgCh = make(chan *config.IntfCfg, VRRP_INTF_CONFIG_CH_SIZE)
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

func (svr *VrrpServer) VrrpStartServer(paramsDir string) {
	svr.OSSignalHandle()
	svr.ReadDB()
	svr.InitGlobalDS()
	svr.GetSystemInfo()
	go svr.EventListener()
}

func VrrpNewServer(sPlugin asicdClient.AsicdClientIntf, dmnBase *dmnBase.FSBaseDmn) *VrrpServer {
	svr := &VrrpServer{}
	svr.SwitchPlugin = sPlugin
	svr.dmnBase = dmnBase
	return svr
}
