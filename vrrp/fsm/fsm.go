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

package fsm

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"l3/vrrp/common"
	"l3/vrrp/debug"
	"l3/vrrp/packet"
	"strconv"
	"syscall"
	"time"
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
}

type DecodedInfo struct {
	PktInfo *packet.PacketInfo
}

/* Handle update in vrrp interface configuration + operstate change config
 */
type IntfEvent struct {
	Event     uint8
	Config    *common.IntfCfg
	OperState string
}

type FSM struct {
	Config              *common.IntfCfg            // config attributes like virtual rtr ip, vrid, intfRef, etc...
	pHandle             *pcap.Handle               // Pcap Handler for receiving packets
	PktInfo             *packet.PacketInfo         // fsm will use this packet infor for decode/encode
	ifIndex             int32                      // My own ifIndex
	ipAddr              string                     // My own ip address
	ipType              int                        // indicates whether fsm is for v4 or v6
	VirtualMACAddress   string                     // VRRP MAC aka VMAC
	State               uint8                      // current state in which fsm is running
	previousState       uint8                      // previous state in which fsm was running
	MasterAdverInterval int32                      // The initial value is the same as Advertisement_Interval.
	SkewTime            int32                      // (((256 - priority) * Master_Adver_Interval) / 256)
	MasterDownValue     int32                      // (3 * Master_Adver_Interval) + Skew_time
	AdverTimer          *time.Timer                // Advertisement Timer
	MasterDownTimer     *time.Timer                // Master down timer...used for keep-alives from master
	stateInfo           *common.State              // this is state information for this fsm which will be used for get bulk
	pktCh               chan *PktChannelInfo       // received vrrp packet are pushed on this channel for fsm
	IntfEventCh         chan *IntfEvent            // channel used by VrrpInterface to communicate and update in config or ip interface state
	vipCh               chan *common.VirtualIpInfo // this will be used to bring up/down virtual ip interface
	rxCh                chan struct{}              // inform server to update global rx count
	txCh                chan struct{}              // inform server to update global tx count
	running             bool
	empty               struct{}
}

/************************************************************************************************************
					* FSM EXPOSED API's *
*************************************************************************************************************/

func InitFsm(cfg *common.IntfCfg, l3Info *common.BaseIpInfo, vipCh chan *common.VirtualIpInfo, rxCh chan struct{}, txCh chan struct{}) *FSM {
	debug.Logger.Info(FSM_PREFIX, "Initializing fsm for vrrp interface:", *cfg, "and base l3 interface is:", *l3Info)
	f := FSM{}
	f.Config = cfg
	f.stateInfo = &common.State{}
	f.ipAddr = l3Info.IpAddr
	f.ifIndex = l3Info.IfIndex
	f.ipType = l3Info.IpType
	f.createVirtualMac()
	f.vipCh = vipCh
	f.pktCh = make(chan *PktChannelInfo)
	f.IntfEventCh = make(chan *IntfEvent)
	f.PktInfo = packet.Init()
	f.State = VRRP_INITIALIZE_STATE
	f.previousState = VRRP_UNINITIALIZE_STATE
	f.rxCh = rxCh
	f.txCh = txCh
	var empty struct{}
	f.empty = empty
	return &f
}

// this will be called when fsm is not running and you want to update the configuration of vrrp interface
func (f *FSM) UpdateConfig(cfg *common.IntfCfg) {
	debug.Logger.Info(FSM_PREFIX, "Changing configuration in fsm:", cfg)
	f.Config = cfg
}

func (f *FSM) DeInitFsm() {
	f.exitFsm()
	f.Config = nil
	f.pHandle = nil
	f.PktInfo = nil
	f.AdverTimer = nil
	f.MasterDownTimer = nil
	f.stateInfo = nil
	f.pktCh = nil
	f.IntfEventCh = nil
}

func (f *FSM) StartFsm() {
	f.running = true
	f.initialize()
	debug.Logger.Debug(FSM_PREFIX, "fsm started for interface:", f.Config.IntfRef)
	for {
		debug.Logger.Debug(FSM_PREFIX)
		select {
		case pktCh, ok := <-f.pktCh:
			if ok {
				// handle received packet
				f.processRcvdPkt(pktCh)
			}
		case intfStateEv, ok := <-f.IntfEventCh:
			if ok {
				debug.Logger.Debug(FSM_PREFIX, "routine received interface event, calling handleIntfEvent for intf:", f.Config.IntfRef)
				f.handleIntfEvent(intfStateEv)
				// special Handling by exiting the go routine
				if intfStateEv.Event == TEAR_DOWN {
					debug.Logger.Info(FSM_PREFIX, "exiting fsm go routine for interface:", f.Config.IntfRef)
					f.running = false
					return
				}
			}
		}
	}
}

func (f *FSM) IsRunning() bool {
	return f.running
}

func (f *FSM) GetStateInfo(info *common.State) {
	debug.Logger.Debug(FSM_PREFIX, "get state info request for:", f.Config.IntfRef)
	info.IntfRef = f.Config.IntfRef
	info.Vrid = f.Config.VRID
	if f.running {
		info.OperState = common.STATE_UP
	} else {
		info.OperState = common.STATE_DOWN
	}
	info.IpAddr = f.ipAddr
	info.CurrentFsmState = f.stateInfo.CurrentFsmState
	info.MasterIp = f.stateInfo.MasterIp
	info.AdverRx = f.stateInfo.AdverRx
	info.AdverTx = f.stateInfo.AdverTx
	info.LastAdverRx = f.stateInfo.LastAdverRx
	info.LastAdverTx = f.stateInfo.LastAdverTx
	info.VirtualIp = f.Config.VirtualIPAddr
	info.VirtualRouterMACAddress = f.VirtualMACAddress
	info.AdvertisementInterval = f.Config.AdvertisementInterval
	info.MasterDownTimer = f.MasterDownValue
	debug.Logger.Debug(FSM_PREFIX, "returning info:", *info)
}

/************************************************************************************************************
					* FSM PRIVATE API's *
*************************************************************************************************************/

func (f *FSM) createVirtualMac() {
	if f.ipType == syscall.AF_INET {
		f.VirtualMACAddress = VERSION2_IEEE_MAC_ADDR_PREFIX
	} else if f.ipType == syscall.AF_INET6 {
		f.VirtualMACAddress = VERSION3_IEEE_MAC_ADDR_PREFIX
	}
	vridStr := strconv.FormatInt(int64(f.Config.VRID), 16)
	if len(vridStr) == 1 {
		f.VirtualMACAddress += "0" + vridStr
	} else {
		f.VirtualMACAddress += vridStr
	}
	debug.Logger.Debug("Vmac created for interface:", f.Config.IntfRef, "is:", f.VirtualMACAddress)
}

func getStateName(state uint8) (rv string) {
	switch state {
	case VRRP_INITIALIZE_STATE:
		rv = VRRP_INITIALIZE_STATE_STRING
	case VRRP_MASTER_STATE:
		rv = VRRP_MASTER_STATE_STRING
	case VRRP_BACKUP_STATE:
		rv = VRRP_BACKUP_STATE_STRING
	}

	return rv
}

func (f *FSM) updateRxStInfo(pktInfo *packet.PacketInfo) {
	f.stateInfo.MasterIp = pktInfo.IpAddr
	f.stateInfo.AdverRx++
	f.stateInfo.LastAdverRx = time.Now().String()
	f.stateInfo.CurrentFsmState = getStateName(f.State)
	f.rxCh <- f.empty
}

func (f *FSM) updateTxStInfo() {
	f.stateInfo.MasterIp = f.ipAddr
	f.stateInfo.AdverTx++
	f.stateInfo.LastAdverTx = time.Now().String()
	f.stateInfo.CurrentFsmState = getStateName(f.State)
	f.txCh <- f.empty
}

func (f *FSM) receivePkt() {
	packetSource := gopacket.NewPacketSource(f.pHandle, f.pHandle.LinkType())
	in := packetSource.Packets()
	ifName := f.Config.IntfRef
	for {
		select {
		case pkt, ok := <-in:
			if !ok {
				debug.Logger.Debug("Pcap closed for interface:", ifName, "exiting RX go routine")
				return
			}
			f.pktCh <- &PktChannelInfo{
				pkt: pkt,
			}
		}
	}
}

func (f *FSM) initPktListener() (err error) {
	ifName := f.Config.IntfRef
	if f.pHandle == nil {
		debug.Logger.Debug(FSM_PREFIX, "initPktListener for interface:", ifName)
		f.pHandle, err = pcap.OpenLive(ifName, VRRP_SNAPSHOT_LEN, VRRP_PROMISCOUS_MODE, VRRP_TIMEOUT)
		if err != nil {
			debug.Logger.Err("Creating Pcap Handle for l3 interface:", ifName, "failed with error:", err)
			return err
		}
		filter := VRRP2_BPF_FILTER
		if f.Config.IpType == syscall.AF_INET6 {
			filter = VRRP3_BPF_FILTER
		}
		err = f.pHandle.SetBPFFilter(filter)
		if err != nil {
			debug.Logger.Err("Setting filter:", filter, "for l3 interface:", ifName, "failed with error:", err)
			return err
		}
		debug.Logger.Debug(FSM_PREFIX, "Pcap created go start Receiving Vrrp Packets")
		// if everything is success then only start receiving packets
		go f.receivePkt()
	}
	return nil
}

func (f *FSM) deInitPktListener() {
	if f.pHandle != nil {
		debug.Logger.Info(FSM_PREFIX, "deInitPktListener for interface:", f.Config.IntfRef, "vrid:", f.Config.VRID)
		f.pHandle.Close()
		f.pHandle = nil
	}
}

func (f *FSM) processRcvdPkt(pktCh *PktChannelInfo) {
	pktInfo := f.PktInfo.Decode(pktCh.pkt, f.Config.IpType)
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
		if f.ipAddr == hdr.IpAddr[i].String() {
			debug.Logger.Err("Header payload ip address is same as my own ip address, FSM INFO ----> intf:",
				f.Config.IntfRef, "ipAddr:", f.ipAddr)
			return
		}
	}
	f.handleDecodedPkt(&DecodedInfo{PktInfo: pktInfo})
}

func (f *FSM) send(pktInfo *packet.PacketInfo) {
	pkt := f.PktInfo.Encode(pktInfo)
	if f.pHandle != nil {
		err := f.pHandle.WritePacketData(pkt)
		if err != nil {
			debug.Logger.Err(FSM_PREFIX, "Writing packet failed for interface:", f.Config.IntfRef)
			return
		}
		f.updateTxStInfo()
	}
}

func (f *FSM) getPacketInfo() *packet.PacketInfo {
	pktInfo := &packet.PacketInfo{
		Version:      f.Config.Version,
		Vrid:         uint8(f.Config.VRID),
		Priority:     uint8(f.Config.Priority), //VRRP_IGNORE_PRIORITY,
		AdvertiseInt: uint16(f.Config.AdvertisementInterval),
		VirutalMac:   f.VirtualMACAddress,
		IpType:       f.Config.IpType,
	}
	if f.Config.VirtualIPAddr == "" {
		// If no virtual ip then use interface/router ip address as virtual ip
		pktInfo.Vip = f.ipAddr
	} else {
		pktInfo.Vip = f.Config.VirtualIPAddr
	}
	pktInfo.IpAddr = f.ipAddr
	return pktInfo
}

func (f *FSM) calculateDownValue(advInt int32) {
	debug.Logger.Debug(FSM_PREFIX, "calculateDownValue with advInt:", advInt)
	//(155) + Set Master_Adver_Interval to Advertisement_Interval
	if advInt == common.USE_CONFIG_ADVERTISEMENT {
		f.MasterAdverInterval = f.Config.AdvertisementInterval
	} else {
		f.MasterAdverInterval = advInt
	}
	//(160) + Set the Master_Down_Timer to Master_Down_Interval
	if f.Config.Priority != 0 && f.MasterAdverInterval != 0 {
		f.SkewTime = ((256 - f.Config.Priority) * f.MasterAdverInterval) / 256
	}
	f.MasterDownValue = (3 * f.MasterAdverInterval) + f.SkewTime
	debug.Logger.Debug(FSM_PREFIX, "MasterAdverInterval is:", f.MasterAdverInterval, "SkewTime is:", f.SkewTime,
		"MasterDownValue is:", f.MasterDownValue)
}

func (f *FSM) initialize() {
	f.initPktListener()
	debug.Logger.Debug(FSM_PREFIX, "In Init state deciding next state")
	switch f.Config.Priority {
	case VRRP_MASTER_PRIORITY:
		debug.Logger.Debug(FSM_PREFIX, "Transition To master")
		f.transitionToMaster()
	default:
		debug.Logger.Debug(FSM_PREFIX, "Tranisition to backup")
		f.transitionToBackup(common.USE_CONFIG_ADVERTISEMENT)
	}
}

func (f *FSM) handleDecodedPkt(decodeInfo *DecodedInfo) {
	//debug.Logger.Debug(FSM_PREFIX, "Processing Decoded Packet")
	switch f.State {
	case VRRP_INITIALIZE_STATE:
		f.initialize()
	case VRRP_BACKUP_STATE:
		f.backup(decodeInfo)
	case VRRP_MASTER_STATE:
		f.master(decodeInfo)
	}
}

/*
 * VRRP_BACKUP_STATE
 *  (345) - If a Shutdown event is received, then:
 *  (350) + Cancel the Master_Down_Timer
 *  (355) + Transition to the {Initialize} state
 * VRRP_MASTER_STATE
 *  (655) - If a Shutdown event is received, then:
 *  (660) + Cancel the Adver_Timer
 *  (665) + Send an ADVERTISEMENT with Priority = 0
 *  (670) + Transition to the {Initialize} state
 *  (675) -endif // shutdown recv
 */
func (f *FSM) stateDownEvent() {
	debug.Logger.Debug(FSM_PREFIX, "handling state down event for:", *f.Config)
	f.stopMasterDownTimer()
	f.stopMasterAdverTimer()

	if f.State == VRRP_MASTER_STATE {
		pkt := f.getPacketInfo()
		pkt.Priority = VRRP_MASTER_DOWN_PRIORITY
		f.send(pkt)
	}
	f.previousState = f.State
	f.State = VRRP_INITIALIZE_STATE
	f.deInitPktListener()
	// remove the virtual ip if state down event
	f.updateVirtualIP(false /*disable*/)
}

func (f *FSM) stateUpEvent() {
	// during state up event move to initialization
	f.initialize()
}

func (f *FSM) exitFsm() {
	f.stateDownEvent()
}

func (f *FSM) handleIntfEvent(intfEvent *IntfEvent) {
	debug.Logger.Info("Handling vrrp interface config update event:", intfEvent.Event)
	switch intfEvent.Event {
	case STATE_CHANGE:
		debug.Logger.Info(FSM_PREFIX, "fsm received state change event", intfEvent.OperState)
		switch intfEvent.OperState {
		case common.STATE_DOWN:
			f.stateDownEvent()
		case common.STATE_UP:
			f.stateUpEvent()
		}
	case TEAR_DOWN:
		// special case...
		debug.Logger.Info(FSM_PREFIX, "Tear down fsm for:", f.Config.IntfRef, "vrid:", f.Config.VRID)
		f.exitFsm()
	}
}

func (f *FSM) updateVirtualIP(enable bool) {
	// Set Sub-intf state up and send out garp via linux stack
	f.vipCh <- &common.VirtualIpInfo{
		IntfRef: f.Config.IntfRef,
		IpAddr:  f.Config.VirtualIPAddr,
		MacAddr: f.VirtualMACAddress,
		Enable:  enable,
		Version: f.Config.Version,
		IpType:  f.Config.IpType,
	}
}

// vrrp states
const (
	VRRP_UNINITIALIZE_STATE = iota
	VRRP_INITIALIZE_STATE
	VRRP_BACKUP_STATE
	VRRP_MASTER_STATE
)

const (
	VRRP_MASTER_PRIORITY          = 255
	VRRP_IGNORE_PRIORITY          = 65535
	VRRP_MASTER_DOWN_PRIORITY     = 0
	VRRP_INITIALIZE_STATE_STRING  = "Initialize"
	VRRP_BACKUP_STATE_STRING      = "Backup"
	VRRP_MASTER_STATE_STRING      = "Master"
	VRRP_SNAPSHOT_LEN             = 1024
	VRRP_PROMISCOUS_MODE          = false
	VRRP_TIMEOUT                  = 1 // in seconds
	VRRP2_BPF_FILTER              = "ip host " + packet.VRRP_V4_GROUP_IP
	VRRP3_BPF_FILTER              = "ip host " + packet.VRRP_V6_GROUP_IP
	VRRP_MAC_MASK                 = "ff:ff:ff:ff:ff:ff"
	FSM_PREFIX                    = "FSM ------> "
	VERSION2_IEEE_MAC_ADDR_PREFIX = "00-00-5E-00-01-"
	VERSION3_IEEE_MAC_ADDR_PREFIX = "00-00-5E-00-02-"
)

const (
	_ = iota
	STATE_CHANGE
	TEAR_DOWN
)
