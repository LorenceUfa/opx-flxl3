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
	"asicdServices"
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"l3/vrrp/packet"
	"net"
	"strconv"
	"strings"
	"time"
)

// vrrp states
const (
	VRRP_UNINITIALIZE_STATE = iota
	VRRP_INITIALIZE_STATE
	VRRP_BACKUP_STATE
	VRRP_MASTER_STATE
)

const (
	VRRP_MASTER_PRIORITY      = 255
	VRRP_IGNORE_PRIORITY      = 65535
	VRRP_MASTER_DOWN_PRIORITY = 0

	FSM_PREFIX = "FSM ------> "
)

/*
	0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                    IPv4 Fields or IPv6 Fields                 |
	...                                                             ...
	|                                                               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|Version| Type  | Virtual Rtr ID|   Priority    |Count IPvX Addr|
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|(rsvd) |     Max Adver Int     |          Checksum             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                                                               |
	+                                                               +
	|                       IPvX Address(es)                        |
	+                                                               +
	+                                                               +
	+                                                               +
	+                                                               +
	|                                                               |
	+                                                               +
	|                                                               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			   +---------------+
		+--------->|               |<-------------+
		|          |  Initialize   |              |
		|   +------|               |----------+   |
		|   |      +---------------+          |   |
		|   |                                 |   |
		|   V                                 V   |
	+---------------+                       +---------------+
	|               |---------------------->|               |
	|    Master     |                       |    Backup     |
	|               |<----------------------|               |
	+---------------+                       +---------------+

*/

type PktChannelInfo struct {
	pkt gopacket.Packet
	/*
		key KeyInfo
			key     string
			IfIndex int32
	*/
}

type FsmStateInfo struct {
	PktInfo *packet.PacketInfo
	//Hdr *packet.Header
}

type FSM struct {
	// config attributes like virtual rtr ip, vrid, intfRef, etc...
	Config *config.IntfCfg
	// My own ip address
	IpAddr string
	// My own ifIndex
	IfIndex int32
	// Pcap Handler for receiving packets
	pHandle *pcap.Handle
	// VRRP MAC aka VMAC
	VirtualRouterMACAddress string

	State uint8

	StateCh *StateInfo

	pktCh *PktChannelInfo

	PktInfo *PacketInfo

	fsmStCh *FsmStateInfo

	txPktCh *packet.PacketInfo

	// Advertisement Timer
	AdverTimer *time.Timer

	// The initial value is the same as Advertisement_Interval.
	MasterAdverInterval int32

	// (((256 - priority) * Master_Adver_Interval) / 256)
	SkewTime int32

	// (3 * Master_Adver_Interval) + Skew_time
	MasterDownValue int32

	MasterDownTimer *time.Timer

	/*
		MasterDownLock  *sync.RWMutex

		// State Name
		//StateName string
		// Lock to read current state of vrrp object
		//StateNameLock *sync.RWMutex
		// Vrrp State Lock for each IfIndex + VRID
		//StateInfo     VrrpGlobalStateInfo
		//StateInfoLock *sync.RWMutex
	*/
}

func InitFsm(cfg *config.IntfCfg, l3Info *L3Intf, stCh *StateInfo) *FSM {
	f := &FSM{}
	f.Config = cfg
	f.IpAddr = l3Info.IpAddr
	f.IfIndex = l3Info.IfIndex
	f.VirtualRouterMACAddress = createVirtualMac(cfg.VRID)
	f.StateCh = stCh
	f.pktCh = make(chan *PktChannelInfo)
	f.fsmStCh = make(chan *FsmStateInfo)
	f.txPktCh = make(chan *packet.PacketInfo)
	f.PktInfo = packet.Init()
	go f.StartFsm()
	f.InitPacketListener()
	f.State = VRRP_INITIALIZE_STATE
	return f
}

func createVirtualMac(vrid int32) (vmac string) {
	if vrid < 10 {
		vmac = VRRP_IEEE_MAC_ADDR + "0" + strconv.Itoa(int(vrid))

	} else {
		vmac = VRRP_IEEE_MAC_ADDR + strconv.Itoa(int(vrid))
	}
	return vmac
}

func (f *FSM) ReceiveVrrpPackets() {
	pHandle := f.pHandle
	pktCh := f.pktCh
	packetSource := gopacket.NewPacketSource(pHandle, pHandle.LinkType())
	in := packetSource.Packets()
	for {
		select {
		case pkt, ok := <-in:
			if !ok {
				debug.Logger.Debug("Pcap closed for interface:", intf.L3.IfName, "exiting RX go routine")
				return
			}
			pktCh <- &PktChannelInfo{
				pkt: pkt,
				/*
					key: KeyInfo{
						IntfRef: intf.L3.IfName,
						VRID:    intf.Config.VRID,
					},
				*/
			}
		}
	}
}

func (f *FSM) InitPacketListener() error {
	var err error
	pHandle := f.pHandle
	ifName := f.Config.IntfRef
	if pHandle == nil {
		pHandle, err = pcap.OpenLive(ifName, VRRP_SNAPSHOT_LEN, VRRP_PROMISCOUS_MODE, VRRP_TIMEOUT)
		if err != nil {
			debug.Logger.Err("Creating Pcap Handle for l3 interface:", ifName, "failed with error:", err)
			return
		}

		err = pHandle.SetBPFFilter(VRRP_BPF_FILTER)
		if err != nil {
			debug.Logger.Err("Setting filter:", VRRP_BPF_FILTER, "for l3 interface:", ifName, "failed with error:", err)
			return err
		}

		// if everything is success then only start receiving packets
		go f.ReceiveVrrpPackets()
	}
}

func (f *FSM) ProcessRcvdPkt(pktCh *config.PktChannelInfo) {
	pktInfo := f.PktInfo.Decode(pktCh.pkt, f.Config.Version)
	if pktInfo == nil {
		debug.Logger.Err("Decoding Vrrp Header Failed")
		return
	}
	hdr := pktInfo.Hdr
	for i := 0; i < int(hdr.CountIPAddr); i++ {
		/* If Virtual Ip is not configured then check whether the ip
		 * address of router/interface is not same as the received
		 * Virtual Ip Addr
		 */
		if f.IpAddr == hdr.IpAddr[i].String() {
			debug.Logger.Err("Header payload ip address is same as my own ip address, FSM INFO ----> intf:",
				f.Config.IntfRef, "ipAddr:", f.IpAddr)
			return
		}
	}
	f.fsmStCh <- &FsmStateInfo{
		//Hdr: hdr,
		PktInfo: pktInfo,
	}
}

func (f *FSM) SendPkt(pktInfo *packet.PacketInfo) {
	pkt := f.PktInfo.Encode(pktInfo)
	if f.pHandle != nil {
		err := f.pHandle.WritePacketData(pkt)
		if err != nil {
			debug.Logger.Err(FSM_PREFIX, "Writing packet failed for interface:", f.Config.IntfRef)
		}
	}
}

func (f *FSM) getPacketInfo() *PacketInfo {
	pktInfo := &packet.PacketInfo{
		Version:      f.Config.Version,
		Vrid:         f.Config.VRID,
		Priority:     VRRP_IGNORE_PRIORITY,
		AdvertiseInt: f.Config.AdvertisementInterval,
		VirutalMac:   f.VirtualRouterMACAddress,
	}
	if f.Config.VirtualIPAddr == "" {
		// If no virtual ip then use interface/router ip address as virtual ip
		pktInfo.IpAddr = f.IpAddr
	} else {
		pktInfo.IpAddr = f.Config.VirtualIPAddr
	}
	return pktInfo
}

func (f *FSM) StartMasterAdverTimer() {
	if f.AdverTimer != nil {
		f.AdverTimer.Reset(time.Duration(f.Config.AdvertisementInterval) * time.Second)
	} else {
		var SendMasterAdveristement_func func()
		SendMasterAdveristement_func = func() {
			// Send advertisment every time interval expiration
			f.SendPkt(f.getPacketInfo())
			f.AdverTimer.Reset(time.Duration(f.Config.AdvertisementInterval) * time.Second)
		}
		debug.Logger.Debug("Setting Master Advertisement Timer to:", f.Config.AdvertisementInterval)
		f.AdverTimer = time.AfterFunc(time.Duration(f.Config.AdvertisementInterval), SendMasterAdveristement_func)
	}

	/*
		var timerCheck_func func()
		timerCheck_func = func() {
			// Send advertisment every time interval expiration
			svr.vrrpTxPktCh <- VrrpTxChannelInfo{
				key:      key,
				priority: VRRP_IGNORE_PRIORITY,
			}
			<-svr.vrrpPktSend
			gblInfo, exists := svr.vrrpGblInfo[key]
			if !exists {
				svr.logger.Err("Gbl Config for " + key + " doesn't exists")
				return
			}
			gblInfo.AdverTimer.Reset(
				time.Duration(gblInfo.IntfConfig.AdvertisementInterval) *
					time.Second)
			svr.vrrpGblInfo[key] = gblInfo
		}
		gblInfo, exists := svr.vrrpGblInfo[key]
		if exists {
			svr.logger.Info(fmt.Sprintln("setting adver timer to",
				gblInfo.IntfConfig.AdvertisementInterval))
			// Set Timer expire func...
			gblInfo.AdverTimer = time.AfterFunc(
				time.Duration(gblInfo.IntfConfig.AdvertisementInterval)*time.Second,
				timerCheck_func)
			// (145) + Transition to the {Master} state
			gblInfo.StateNameLock.Lock()
			gblInfo.StateName = VRRP_MASTER_STATE
			gblInfo.StateNameLock.Unlock()
			svr.vrrpGblInfo[key] = gblInfo
		}
	*/
}

func (f *FSM) StopMasterAdverTimer() {
	if f.AdverTimer != nil {
		f.AdverTimer.Stop()
		f.AdverTimer = nil
	}
}

func (f *FSM) StartMasterDownTimer(key string) {
	if f.MasterDownTimer != nil {
		gblInfo.MasterDownTimer.Reset(time.Duration(f.MasterDownValue) * time.Second)
	} else {
		var MasterDownTimer_func func()
		// On Timer expiration we will transition to master
		MasterDownTimer_func = func() {
			debug.Logger.Info(FSM_PREFIX, "master down timer expired..transition to Master")
			// @TODO: do timer expiry handling here
			svr.VrrpTransitionToMaster(key, "Master Down Timer expired")
		}
		debug.logger.Info("setting down timer to", f.MasterDownValue)
		// Set Timer expire func...
		gblInfo.MasterDownTimer = time.AfterFunc(time.Duration(f.MasterDownValue)*time.Second,
			MasterDownTimer_func)
	}
	gblInfo.StateNameLock.Lock()
	gblInfo.StateName = VRRP_BACKUP_STATE
	gblInfo.StateNameLock.Unlock()
	svr.vrrpGblInfo[key] = gblInfo
}

func (f *FSM) CalculateDownValue() {
	//(155) + Set Master_Adver_Interval to Advertisement_Interval
	f, MasterAdverInterval = f.Config.AdvertisementInterval
	//(160) + Set the Master_Down_Timer to Master_Down_Interval
	if f.Config.Priority != 0 && f.Config.MasterAdverInterval != 0 {
		f.Config.SkewTime = ((256 - f.Config.Priority) * f.Config.MasterAdverInterval) / 256
	}
	f.MasterDownValue = (3 * f.MasterAdverInterval) + f.SkewTime
}

func (f *FSM) TransitionToMaster() {
	pktInfo := f.getPacketInfo()
	// (110) + Send an ADVERTISEMENT
	f.SendPkt(pktInfo)
	// (145) + Transition to the {Master} state
	f.State = VRRP_MASTER_STATE
	// @TODO : Set Sub-intf state up and send out garp via linux stack
	// svr.VrrpUpdateSubIntf(gblInfo, true /*configure or set*/ //)

	// (140) + Set the Adver_Timer to Advertisement_Interval
	// Start Advertisement Timer
	f.StartMasterAdverTimer()
}

func (f *FSM) TransitionToBackup() {
	debug.Logger.Debug(FSM_PREFIX, "advertisement timer to be used in backup state for",
		"calculating master down timer is ", f.Config.AdvertisementInterval)
	// @TODO: Bring Down Sub-Interface
	//	svr.VrrpUpdateSubIntf(gblInfo, false /*configure or set*/)

	// Re-Calculate Down timer value
	f.CalculateDownValue()
	f.StartMasterDownTimer()
	//(165) + Transition to the {Backup} state
	f.State = VRRP_BACKUP_STATE
	//svr.VrrpUpdateStateInfo(key, reason, VRRP_BACKUP_STATE)
}

func (f *FSM) Initialize() {
	debug.Logger.Debug(FSM_PREFIX, "In Init state deciding next state")

	switch f.Config.Priority {
	case VRRP_MASTER_PRIORITY:
		f.TransitionToMaster()
	default:
		f.TransitionToBackup()
	}
}

func (f *FSM) MasterState(stInfo *FsmStateInfo) {
	debug.Logger.Debug(FSM_PREFIX, "In Master State Handling Fsm Info:", *stInfo)
	pktInfo := stInfo.PktInfo
	hdr := pktInfo.Hdr
	/* // @TODO:
	   (645) - MUST forward packets with a destination link-layer MAC
	   address equal to the virtual router MAC address.

	   (650) - MUST accept packets addressed to the IPvX address(es)
	   associated with the virtual router if it is the IPvX address owner
	   or if Accept_Mode is True.  Otherwise, MUST NOT accept these
	   packets.
	*/
	//  (700) - If an ADVERTISEMENT is received, then:
	//	 (705) -+ If the Priority in the ADVERTISEMENT is zero, then:
	if hdr.Priority == VRRP_MASTER_DOWN_PRIORITY {
		// (710) -* Send an ADVERTISEMENT
		debug.Logger.Debug(FSM_PREFIX, "Priority in the ADVERTISEMENT is zero, then: Send an ADVERTISEMENT")
		f.SendPkt(f.getPacketInfo())
		// (715) -* Reset the Adver_Timer to Advertisement_Interval
		f.StartMasterAdverTimer()
	} else { // (720) -+ else // priority was non-zero
		/*     (725) -* If the Priority in the ADVERTISEMENT is greater than the local Priority,
		*      (730) -* or
		*      (735) -* If the Priority in the ADVERTISEMENT is equal to
		*               the local Priority and the primary IPvX Address of the
		*	        sender is greater than the local primary IPvX Address, then:
		 */
		if int32(hdr.Priority) > f.Config.Priority ||
			(int32(hdr.Priority) == f.Config.Priority &&
				bytes.Compare(net.ParseIP(pktInfo.IpAddr), net.ParseIP(f.IpAddr)) > 0) {
			// (740) -@ Cancel Adver_Timer
			f.StopMasterAdverTimer()
			svr.VrrpTransitionToBackup(key, int32(vrrpHdr.MaxAdverInt),
				"Remote Priority is higher OR (priority are equal AND remote ip is higher than local ip)")
		} else { // new Master logic
			// Discard Advertisement
			return
		} // endif new Master Detected
	} // end if was priority zero
	// end for Advertisemtn received over the channel
	// end MASTER STATE
}

func (f *FSM) ProcessStateInfo(fsmStInfo *FsmStateInfo) {
	debug.Logger.Debug(FSM_PREFIX, "Processing State Information")
	switch f.State {
	case VRRP_INITIALIZE_STATE:
		f.Initialize()
	case VRRP_BACKUP_STATE:
	case VRRP_MASTER_STATE:
		f.MasterState(fsmStInfo)
	}
	/*
		key := fsmObj.key
		pktInfo := fsmObj.inPkt
		pktHdr := fsmObj.vrrpHdr
		gblInfo, exists := svr.vrrpGblInfo[key]
		if !exists {
			svr.logger.Err("No entry found ending fsm")
			return
		}
		gblInfo.StateNameLock.Lock()
		currentState := gblInfo.StateName
		gblInfo.StateNameLock.Unlock()
		switch currentState {
		case VRRP_INITIALIZE_STATE:
			svr.VrrpInitState(key)
		case VRRP_BACKUP_STATE:
			svr.VrrpBackupState(pktInfo, pktHdr, key)
		case VRRP_MASTER_STATE:
			svr.VrrpMasterState(pktInfo, pktHdr, key)
		default: // VRRP_UNINTIALIZE_STATE
			svr.logger.Info("No Ip address and hence no need for fsm")
		}
	*/
}

func (f *FSM) StartFsm() {
	for {
		select {
		case pktCh, ok := <-f.pktCh:
			if ok {
				f.ProcessRcvdPkt(pktCh)
				// handle received packet
			}
		case fsmStInfo, ok := <-f.fsmStCh:
			if ok {
				f.ProcessStateInfo(fsmStInfo)
			}
		}
	}
}

/*
type VrrpFsmIntf interface {
	VrrpFsmStart(fsmObj VrrpFsm)
	VrrpCreateObject(gblInfo VrrpGlobalInfo) (fsmObj VrrpFsm)
	VrrpInitState(key string)
	VrrpBackupState(inPkt gopacket.Packet, vrrpHdr *VrrpPktHeader, key string)
	VrrpMasterState(inPkt gopacket.Packet, vrrpHdr *VrrpPktHeader, key string)
	VrrpTransitionToMaster(key string, reason string)
	VrrpTransitionToBackup(key string, AdvertisementInterval int32, reason string)
	VrrpHandleIntfUpEvent(IfIndex int32)
	VrrpHandleIntfShutdownEvent(IfIndex int32)
}
*/
/*
			   +---------------+
		+--------->|               |<-------------+
		|          |  Initialize   |              |
		|   +------|               |----------+   |
		|   |      +---------------+          |   |
		|   |                                 |   |
		|   V                                 V   |
	+---------------+                       +---------------+
	|               |---------------------->|               |
	|    Master     |                       |    Backup     |
	|               |<----------------------|               |
	+---------------+                       +---------------+

*/

func (svr *VrrpServer) VrrpCreateObject(gblInfo VrrpGlobalInfo) (fsmObj VrrpFsm) {
	vrrpHeader := VrrpPktHeader{
		Version:       VRRP_VERSION2,
		Type:          VRRP_PKT_TYPE_ADVERTISEMENT,
		VirtualRtrId:  uint8(gblInfo.IntfConfig.VRID),
		Priority:      uint8(gblInfo.IntfConfig.Priority),
		CountIPv4Addr: 1, // FIXME for more than 1 vip
		Rsvd:          VRRP_RSVD,
		MaxAdverInt:   uint16(gblInfo.IntfConfig.AdvertisementInterval),
		CheckSum:      VRRP_HDR_CREATE_CHECKSUM,
	}

	return VrrpFsm{
		vrrpHdr: &vrrpHeader,
	}
}

/*
 * This API will create config object with MacAddr and configure....
 * Configure will enable/disable the link...
 */
/*
func (svr *VrrpServer) VrrpUpdateSubIntf(gblInfo VrrpGlobalInfo, configure bool) {
	vip := gblInfo.IntfConfig.VirtualIPv4Addr
	if !strings.Contains(vip, "/") {
		vip = vip + "/32"
	}
	config := asicdServices.SubIPv4Intf{
		IpAddr:  vip,
		IntfRef: strconv.Itoa(int(gblInfo.IntfConfig.IfIndex)),
		Enable:  configure,
		MacAddr: gblInfo.VirtualRouterMACAddress,
	}
	svr.logger.Info(fmt.Sprintln("updating sub interface config obj is", config))
	/*
		struct SubIPv4Intf {
			0 1 : string IpAddr
			1 2 : i32 IfIndex
			2 3 : string Type
			3 4 : string MacAddr
			4 5 : bool Enable
		}
*/
/*
	var attrset []bool
	// The len of attrset is set to 5 for 5 elements in the object...
	// if no.of elements changes then index for mac address and enable needs
	// to change..
	attrset = make([]bool, 5)
	elems := len(attrset)
	attrset[elems-1] = true
	if configure {
		attrset[elems-2] = true
	}
	_, err := svr.asicdClient.ClientHdl.UpdateSubIPv4Intf(&config, &config,
		attrset, nil)
	if err != nil {
		svr.logger.Err(fmt.Sprintln("updating sub interface config failed",
			"Error:", err))
	}
	return
}
*/

/*
func (svr *VrrpServer) VrrpUpdateStateInfo(key string, reason string,
	currentSt string) {
	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No entry found ending fsm")
		return
	}
	gblInfo.StateInfoLock.Lock()
	gblInfo.StateInfo.CurrentFsmState = currentSt
	gblInfo.StateNameLock.Lock()
	gblInfo.StateInfo.PreviousFsmState = gblInfo.StateName
	gblInfo.StateNameLock.Unlock()
	gblInfo.StateInfo.ReasonForTransition = reason
	gblInfo.StateInfoLock.Unlock()
	svr.vrrpGblInfo[key] = gblInfo
}

func (svr *VrrpServer) VrrpHandleMasterAdverTimer(key string) {
	var timerCheck_func func()
	timerCheck_func = func() {
		// Send advertisment every time interval expiration
		svr.vrrpTxPktCh <- VrrpTxChannelInfo{
			key:      key,
			priority: VRRP_IGNORE_PRIORITY,
		}
		<-svr.vrrpPktSend
		gblInfo, exists := svr.vrrpGblInfo[key]
		if !exists {
			svr.logger.Err("Gbl Config for " + key + " doesn't exists")
			return
		}
		gblInfo.AdverTimer.Reset(
			time.Duration(gblInfo.IntfConfig.AdvertisementInterval) *
				time.Second)
		svr.vrrpGblInfo[key] = gblInfo
	}
	gblInfo, exists := svr.vrrpGblInfo[key]
	if exists {
		svr.logger.Info(fmt.Sprintln("setting adver timer to",
			gblInfo.IntfConfig.AdvertisementInterval))
		// Set Timer expire func...
		gblInfo.AdverTimer = time.AfterFunc(
			time.Duration(gblInfo.IntfConfig.AdvertisementInterval)*time.Second,
			timerCheck_func)
		// (145) + Transition to the {Master} state
		gblInfo.StateNameLock.Lock()
		gblInfo.StateName = VRRP_MASTER_STATE
		gblInfo.StateNameLock.Unlock()
		svr.vrrpGblInfo[key] = gblInfo
	}
}

*/
/*
func (svr *VrrpServer) VrrpTransitionToMaster(key string, reason string) {
	// (110) + Send an ADVERTISEMENT
	svr.vrrpTxPktCh <- VrrpTxChannelInfo{
		key:      key,
		priority: VRRP_IGNORE_PRIORITY,
	}
	// Wait for the packet to be send out
	<-svr.vrrpPktSend
	// After Advertisment update fsm state info
	svr.VrrpUpdateStateInfo(key, reason, VRRP_MASTER_STATE)

	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No entry found ending fsm")
		return
	}
	// Set Sub-intf state up and send out garp via linux stack
	svr.VrrpUpdateSubIntf(gblInfo, true /*configure or set*/ //)
/*
	// (140) + Set the Adver_Timer to Advertisement_Interval
	// Start Advertisement Timer
	svr.VrrpHandleMasterAdverTimer(key)
}
*/

func (svr *VrrpServer) VrrpHandleMasterDownTimer(key string) {
	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No object for " + key)
		return
	}
	if gblInfo.MasterDownTimer != nil {
		gblInfo.MasterDownLock.Lock()
		gblInfo.MasterDownTimer.Reset(time.Duration(gblInfo.MasterDownValue) *
			time.Second)
		gblInfo.MasterDownLock.Unlock()
	} else {
		var timerCheck_func func()
		// On Timer expiration we will transition to master
		timerCheck_func = func() {
			svr.logger.Info(fmt.Sprintln("master down timer",
				"expired..transition to Master"))
			// do timer expiry handling here
			svr.VrrpTransitionToMaster(key, "Master Down Timer expired")
		}
		svr.logger.Info("initiating master down timer")
		svr.logger.Info(fmt.Sprintln("setting down timer to",
			gblInfo.MasterDownValue))
		// Set Timer expire func...
		gblInfo.MasterDownLock.Lock()
		gblInfo.MasterDownTimer = time.AfterFunc(
			time.Duration(gblInfo.MasterDownValue)*time.Second,
			timerCheck_func)
		gblInfo.MasterDownLock.Unlock()
	}
	//(165) + Transition to the {Backup} state
	gblInfo.StateNameLock.Lock()
	gblInfo.StateName = VRRP_BACKUP_STATE
	gblInfo.StateNameLock.Unlock()
	svr.vrrpGblInfo[key] = gblInfo
}

func (svr *VrrpServer) VrrpCalculateDownValue(AdvertisementInterval int32,
	gblInfo *VrrpGlobalInfo) {
	//(155) + Set Master_Adver_Interval to Advertisement_Interval
	gblInfo.MasterAdverInterval = AdvertisementInterval
	//(160) + Set the Master_Down_Timer to Master_Down_Interval
	if gblInfo.IntfConfig.Priority != 0 && gblInfo.MasterAdverInterval != 0 {
		gblInfo.SkewTime = ((256 - gblInfo.IntfConfig.Priority) *
			gblInfo.MasterAdverInterval) / 256
	}
	gblInfo.MasterDownValue = (3 * gblInfo.MasterAdverInterval) + gblInfo.SkewTime
}

func (svr *VrrpServer) VrrpTransitionToBackup(key string, AdvertisementInterval int32,
	reason string) {
	svr.logger.Info(fmt.Sprintln("advertisement timer to be used in backup state for",
		"calculating master down timer is ", AdvertisementInterval))
	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No entry found ending fsm")
		return
	}
	// Bring Down Sub-Interface
	svr.VrrpUpdateSubIntf(gblInfo, false /*configure or set*/)
	// Re-Calculate Down timer value
	gblInfo.MasterDownLock.Lock()
	svr.VrrpCalculateDownValue(AdvertisementInterval, &gblInfo)
	gblInfo.MasterDownLock.Unlock()
	svr.vrrpGblInfo[key] = gblInfo
	svr.VrrpUpdateStateInfo(key, reason, VRRP_BACKUP_STATE)
	svr.VrrpHandleMasterDownTimer(key)
}

/*
func (svr *VrrpServer) VrrpInitState(key string) {
	svr.logger.Info("in init state decide next state")
	gblInfo, found := svr.vrrpGblInfo[key]
	if !found {
		svr.logger.Err("running info not found, bailing fsm")
		return
	}
	if gblInfo.IntfConfig.Priority == VRRP_MASTER_PRIORITY {
		svr.logger.Info("Transitioning to Master State")
		svr.VrrpTransitionToMaster(key, "Priority is 255")
	} else {
		svr.logger.Info("Transitioning to Backup State")
		// Transition to backup state first
		svr.VrrpTransitionToBackup(key,
			gblInfo.IntfConfig.AdvertisementInterval,
			"Priority is not 255")
	}
}
*/

func (svr *VrrpServer) VrrpBackupState(inPkt gopacket.Packet, vrrpHdr *VrrpPktHeader,
	key string) {
	// @TODO: Handle arp drop...
	// Check dmac address from the inPacket and if it is same discard the packet
	ethLayer := inPkt.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		svr.logger.Err("Not an eth packet?")
		return
	}
	eth := ethLayer.(*layers.Ethernet)
	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No entry found ending fsm")
		return
	}
	if (eth.DstMAC).String() == gblInfo.VirtualRouterMACAddress {
		svr.logger.Err("Dmac is equal to VMac and hence fsm is aborted")
		return
	}
	// MUST NOT accept packets addressed to the IPvX address(es)
	// associated with the virtual router. @TODO: check with Hari
	ipLayer := inPkt.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		svr.logger.Err("Not an ip packet?")
		return
	}
	ipHdr := ipLayer.(*layers.IPv4)
	if (ipHdr.DstIP).String() == gblInfo.IpAddr {
		svr.logger.Err("dst ip is equal to interface ip, drop the packet")
		return
	}

	if vrrpHdr.Type == VRRP_PKT_TYPE_ADVERTISEMENT {
		gblInfo.StateInfoLock.Lock()
		gblInfo.StateInfo.MasterIp = ipHdr.SrcIP.String()
		gblInfo.StateInfo.AdverRx++
		gblInfo.StateInfo.LastAdverRx = time.Now().String()
		gblInfo.StateInfo.CurrentFsmState = gblInfo.StateName
		gblInfo.StateInfoLock.Unlock()
		svr.vrrpGblInfo[key] = gblInfo
		if vrrpHdr.Priority == 0 {
			// Change down Value to Skew time
			gblInfo.MasterDownLock.Lock()
			gblInfo.MasterDownValue = gblInfo.SkewTime
			gblInfo.MasterDownLock.Unlock()
			svr.vrrpGblInfo[key] = gblInfo
			svr.VrrpHandleMasterDownTimer(key)
		} else {
			// local preempt is false
			if gblInfo.IntfConfig.PreemptMode == false {
				// if remote priority is higher update master down
				// timer and move on
				if vrrpHdr.Priority >= uint8(gblInfo.IntfConfig.Priority) {
					gblInfo.MasterDownLock.Lock()
					svr.VrrpCalculateDownValue(int32(vrrpHdr.MaxAdverInt),
						&gblInfo)
					gblInfo.MasterDownLock.Unlock()
					svr.vrrpGblInfo[key] = gblInfo
					svr.VrrpHandleMasterDownTimer(key)
				} else {
					// Do nothing.... same as discarding packet
					svr.logger.Info("Discarding advertisment")
					return
				}
			} else { // local preempt is true
				if vrrpHdr.Priority >= uint8(gblInfo.IntfConfig.Priority) {
					// Do nothing..... same as discarding packet
					svr.logger.Info("Discarding advertisment")
					return
				} else { // Preempt is true... need to take over
					// as master
					svr.VrrpTransitionToMaster(key,
						"Preempt is true and local Priority is higher than remote")
				}
			} // endif preempt test
		} // endif was priority zero
	} // endif was advertisement received
	// end BACKUP STATE
}

/*
func (svr *VrrpServer) VrrpMasterState(inPkt gopacket.Packet, vrrpHdr *VrrpPktHeader,
	key string) {
	/* // @TODO:
	   (645) - MUST forward packets with a destination link-layer MAC
	   address equal to the virtual router MAC address.

	   (650) - MUST accept packets addressed to the IPvX address(es)
	   associated with the virtual router if it is the IPvX address owner
	   or if Accept_Mode is True.  Otherwise, MUST NOT accept these
	   packets.
*/
/*
	if vrrpHdr.Priority == VRRP_MASTER_DOWN_PRIORITY {
		svr.vrrpTxPktCh <- VrrpTxChannelInfo{
			key:      key,
			priority: VRRP_IGNORE_PRIORITY,
		}
		<-svr.vrrpPktSend
		svr.VrrpHandleMasterAdverTimer(key)
	} else {
		ipLayer := inPkt.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			svr.logger.Err("Not an ip packet?")
			return
		}
		ipHdr := ipLayer.(*layers.IPv4)
		gblInfo, exists := svr.vrrpGblInfo[key]
		if !exists {
			svr.logger.Err("No entry found ending fsm")
			return
		}
		if int32(vrrpHdr.Priority) > gblInfo.IntfConfig.Priority ||
			(int32(vrrpHdr.Priority) == gblInfo.IntfConfig.Priority &&
				bytes.Compare(ipHdr.SrcIP,
					net.ParseIP(gblInfo.IpAddr)) > 0) {
			if gblInfo.AdverTimer != nil {
				gblInfo.AdverTimer.Stop()
			}
			svr.vrrpGblInfo[key] = gblInfo
			svr.VrrpTransitionToBackup(key, int32(vrrpHdr.MaxAdverInt),
				"Remote Priority is higher OR (priority are equal AND remote ip is higher than local ip)")
		} else { // new Master logic
			// Discard Advertisement
			return
		} // endif new Master Detected
	} // end if was priority zero
	// end for Advertisemtn received over the channel
	// end MASTER STATE
}
*/

func (svr *VrrpServer) VrrpFsmStart(fsmObj VrrpFsm) {
	key := fsmObj.key
	pktInfo := fsmObj.inPkt
	pktHdr := fsmObj.vrrpHdr
	gblInfo, exists := svr.vrrpGblInfo[key]
	if !exists {
		svr.logger.Err("No entry found ending fsm")
		return
	}
	gblInfo.StateNameLock.Lock()
	currentState := gblInfo.StateName
	gblInfo.StateNameLock.Unlock()
	switch currentState {
	case VRRP_INITIALIZE_STATE:
		svr.VrrpInitState(key)
	case VRRP_BACKUP_STATE:
		svr.VrrpBackupState(pktInfo, pktHdr, key)
	case VRRP_MASTER_STATE:
		svr.VrrpMasterState(pktInfo, pktHdr, key)
	default: // VRRP_UNINTIALIZE_STATE
		svr.logger.Info("No Ip address and hence no need for fsm")
	}
}

/*
 * During a shutdown event stop timers will be called and we will cancel master
 * down timer and transition to initialize state
 */
func (svr *VrrpServer) VrrpStopTimers(IfIndex int32) {
	for _, key := range svr.vrrpIntfStateSlice {
		splitString := strings.Split(key, "_")
		// splitString = { IfIndex, VRID }
		ifindex, _ := strconv.Atoi(splitString[0])
		if int32(ifindex) != IfIndex {
			// Key doesn't match
			continue
		}
		// If IfIndex matches then use that key and stop the timer for
		// that VRID
		gblInfo, found := svr.vrrpGblInfo[key]
		if !found {
			svr.logger.Err("No entry found for Ifindex:" +
				splitString[0] + " VRID:" + splitString[1])
			return
		}
		svr.logger.Info("Stopping Master Down Timer for Ifindex:" +
			splitString[0] + " VRID:" + splitString[1])
		if gblInfo.MasterDownTimer != nil {
			gblInfo.MasterDownTimer.Stop()
		}
		svr.logger.Info("Stopping Master Advertisemen Timer for Ifindex:" +
			splitString[0] + " VRID:" + splitString[1])
		if gblInfo.AdverTimer != nil {
			gblInfo.AdverTimer.Stop()
		}
		// If state is Master then we need to send an advertisement with
		// priority as 0
		gblInfo.StateNameLock.RLock()
		state := gblInfo.StateName
		gblInfo.StateNameLock.RUnlock()
		if state == VRRP_MASTER_STATE {
			svr.vrrpTxPktCh <- VrrpTxChannelInfo{
				key:      key,
				priority: VRRP_MASTER_DOWN_PRIORITY,
			}
			<-svr.vrrpPktSend
		}
		// Transition to Init State
		gblInfo.StateNameLock.Lock()
		gblInfo.StateName = VRRP_INITIALIZE_STATE
		gblInfo.StateNameLock.Unlock()
		svr.vrrpGblInfo[key] = gblInfo
		svr.logger.Info(fmt.Sprintln("VRID:", gblInfo.IntfConfig.VRID,
			" transitioned to INIT State"))
	}
}

func (svr *VrrpServer) VrrpHandleIntfShutdownEvent(IfIndex int32) {
	svr.VrrpStopTimers(IfIndex)
}

func (svr *VrrpServer) VrrpHandleIntfUpEvent(IfIndex int32) {
	for _, key := range svr.vrrpIntfStateSlice {
		splitString := strings.Split(key, "_")
		// splitString = { IfIndex, VRID }
		ifindex, _ := strconv.Atoi(splitString[0])
		if int32(ifindex) != IfIndex {
			// Key doesn't match
			continue
		}
		// If IfIndex matches then use that key and stop the timer for
		// that VRID
		gblInfo, found := svr.vrrpGblInfo[key]
		if !found {
			svr.logger.Err("No entry found for Ifindex:" +
				splitString[0] + " VRID:" + splitString[1])
			return
		}

		svr.logger.Info(fmt.Sprintln("Intf State Up Notification",
			" restarting the fsm event for VRID:", gblInfo.IntfConfig.VRID))
		svr.vrrpFsmCh <- VrrpFsm{
			key: key,
		}
	}
}
