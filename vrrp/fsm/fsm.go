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
	"l3/vrrp/config"
	"l3/vrrp/debug"
	"l3/vrrp/packet"
	"strconv"
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
	Config    *config.IntfCfg
	OperState string
}

type FSM struct {
	Config                  *config.IntfCfg    // config attributes like virtual rtr ip, vrid, intfRef, etc...
	pHandle                 *pcap.Handle       // Pcap Handler for receiving packets
	PktInfo                 *packet.PacketInfo // fsm will use this packet infor for decode/encode
	IfIndex                 int32              // My own ifIndex
	IpAddr                  string             // My own ip address
	VirtualRouterMACAddress string             // VRRP MAC aka VMAC
	State                   uint8              // current state in which fsm is running
	MasterAdverInterval     int32              // The initial value is the same as Advertisement_Interval.
	SkewTime                int32              // (((256 - priority) * Master_Adver_Interval) / 256)
	MasterDownValue         int32              // (3 * Master_Adver_Interval) + Skew_time
	AdverTimer              *time.Timer        // Advertisement Timer
	MasterDownTimer         *time.Timer        // Master down timer...used for keep-alives from master
	stateInfo               *config.State      // this is state information for this fsm which will be used for get bulk
	pktCh                   chan *PktChannelInfo
	decodeCh                chan *DecodedInfo
	txPktCh                 chan *packet.PacketInfo
	IntfEventCh             chan *IntfEvent
	vipCh                   chan *config.VirtualIpInfo // this will be used to bring up/down virtual ip interface
	running                 bool
}

func InitFsm(cfg *config.IntfCfg, l3Info *config.BaseIpInfo, vipCh chan *config.VirtualIpInfo) *FSM {
	debug.Logger.Info(FSM_PREFIX, "Initializing fsm for vrrp interface:", *cfg, "and base l3 interface is:", *l3Info)
	f := &FSM{}
	f.Config = cfg
	f.stateInfo = &config.State{}
	f.IpAddr = l3Info.IpAddr
	f.IfIndex = l3Info.IfIndex
	f.VirtualRouterMACAddress = createVirtualMac(cfg.VRID)
	f.vipCh = vipCh
	f.pktCh = make(chan *PktChannelInfo)
	f.decodeCh = make(chan *DecodedInfo)
	f.txPktCh = make(chan *packet.PacketInfo)
	f.IntfEventCh = make(chan *IntfEvent)
	f.PktInfo = packet.Init()
	f.State = VRRP_INITIALIZE_STATE
	return f
}

func createVirtualMac(vrid int32) (vmac string) {
	if vrid < 10 {
		vmac = packet.VRRP_IEEE_MAC_ADDR + "0" + strconv.Itoa(int(vrid))

	} else {
		vmac = packet.VRRP_IEEE_MAC_ADDR + strconv.Itoa(int(vrid))
	}
	return vmac
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

func (f *FSM) IsRunning() bool {
	return f.running
}

func (f *FSM) UpdateRxStateInformation(pktInfo *packet.PacketInfo) {
	f.stateInfo.MasterIp = pktInfo.IpAddr
	f.stateInfo.AdverRx++
	f.stateInfo.LastAdverRx = time.Now().String()
	f.stateInfo.CurrentFsmState = getStateName(f.State)
}

func (f *FSM) UpdateTxStateInformation() {
	f.stateInfo.MasterIp = f.IpAddr
	f.stateInfo.AdverTx++
	f.stateInfo.LastAdverTx = time.Now().String()
	f.stateInfo.CurrentFsmState = getStateName(f.State)
}

func (f *FSM) GetStateInfo(info *config.State) {
	debug.Logger.Debug(FSM_PREFIX, "get state info request for:", f.Config.IntfRef)
	info.IntfRef = f.Config.IntfRef
	info.Vrid = f.Config.VRID
	info.IpAddr = f.IpAddr
	info.CurrentFsmState = f.stateInfo.CurrentFsmState
	info.MasterIp = f.stateInfo.MasterIp
	info.AdverRx = f.stateInfo.AdverRx
	info.AdverTx = f.stateInfo.AdverTx
	info.LastAdverRx = f.stateInfo.LastAdverRx
	info.LastAdverTx = f.stateInfo.LastAdverTx
	info.VirtualIp = f.Config.VirtualIPAddr
	info.VirtualRouterMACAddress = f.VirtualRouterMACAddress
	info.MasterDownTimer = f.MasterDownValue
	debug.Logger.Debug(FSM_PREFIX, "returning info:", *info)
}

func (f *FSM) ReceiveVrrpPackets() {
	packetSource := gopacket.NewPacketSource(f.pHandle, f.pHandle.LinkType())
	in := packetSource.Packets()
	for {
		select {
		case pkt, ok := <-in:
			if !ok {
				debug.Logger.Debug("Pcap closed for interface:", f.Config.IntfRef, "exiting RX go routine")
				return
			}
			f.pktCh <- &PktChannelInfo{
				pkt: pkt,
			}
		}
	}
}

func (f *FSM) InitPacketListener() (err error) {
	ifName := f.Config.IntfRef
	if f.pHandle == nil {
		debug.Logger.Debug(FSM_PREFIX, "InitPacketListener for interface:", ifName)
		f.pHandle, err = pcap.OpenLive(ifName, VRRP_SNAPSHOT_LEN, VRRP_PROMISCOUS_MODE, VRRP_TIMEOUT)
		if err != nil {
			debug.Logger.Err("Creating Pcap Handle for l3 interface:", ifName, "failed with error:", err)
			return err
		}

		err = f.pHandle.SetBPFFilter(VRRP_BPF_FILTER)
		if err != nil {
			debug.Logger.Err("Setting filter:", VRRP_BPF_FILTER, "for l3 interface:", ifName, "failed with error:", err)
			return err
		}
		debug.Logger.Debug("Pcap created go start Receiving Vrrp Packets")
		// if everything is success then only start receiving packets
		go f.ReceiveVrrpPackets()
	}
	return nil
}

func (f *FSM) DeInitPacketListener() {
	if f.pHandle != nil {
		debug.Logger.Info(FSM_PREFIX, "DeInitPacketListener for interface:", f.Config.IntfRef, "vrid:", f.Config.VRID)
		f.pHandle.Close()
		f.pHandle = nil
	}
}

func (f *FSM) ProcessRcvdPkt(pktCh *PktChannelInfo) {
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
	f.decodeCh <- &DecodedInfo{
		PktInfo: pktInfo,
	}
}

func (f *FSM) SendPkt(pktInfo *packet.PacketInfo) {
	pkt := f.PktInfo.Encode(pktInfo)
	if f.pHandle != nil {
		err := f.pHandle.WritePacketData(pkt)
		if err != nil {
			debug.Logger.Err(FSM_PREFIX, "Writing packet failed for interface:", f.Config.IntfRef)
			return
		}
		debug.Logger.Debug(FSM_PREFIX, "updating Tx state information")
		f.UpdateTxStateInformation()
	}
}

func (f *FSM) getPacketInfo() *packet.PacketInfo {
	pktInfo := &packet.PacketInfo{
		Version:      f.Config.Version,
		Vrid:         uint8(f.Config.VRID),
		Priority:     uint8(f.Config.Priority), //VRRP_IGNORE_PRIORITY,
		AdvertiseInt: uint16(f.Config.AdvertisementInterval),
		VirutalMac:   f.VirtualRouterMACAddress,
	}
	if f.Config.VirtualIPAddr == "" {
		// If no virtual ip then use interface/router ip address as virtual ip
		pktInfo.Vip = f.IpAddr
	} else {
		pktInfo.Vip = f.Config.VirtualIPAddr
	}
	pktInfo.IpAddr = f.IpAddr
	return pktInfo
}

func (f *FSM) CalculateDownValue(advInt int32) {
	debug.Logger.Debug(FSM_PREFIX, "CalculateDownValue with advInt:", advInt)
	//(155) + Set Master_Adver_Interval to Advertisement_Interval
	if advInt == config.USE_CONFIG_ADVERTISEMENT {
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

func (f *FSM) Initialize() {
	f.InitPacketListener()
	debug.Logger.Debug(FSM_PREFIX, "In Init state deciding next state")
	switch f.Config.Priority {
	case VRRP_MASTER_PRIORITY:
		debug.Logger.Debug(FSM_PREFIX, "Transition To MasterState")
		f.TransitionToMaster()
	default:
		debug.Logger.Debug(FSM_PREFIX, "Tranisition to BackupState")
		f.TransitionToBackup(config.USE_CONFIG_ADVERTISEMENT)
	}
}

func (f *FSM) HandleDecodedPkt(decodeInfo *DecodedInfo) {
	debug.Logger.Debug(FSM_PREFIX, "Processing Decoded Packet")
	switch f.State {
	case VRRP_INITIALIZE_STATE:
		f.Initialize()
	case VRRP_BACKUP_STATE:
		f.BackupState(decodeInfo)
	case VRRP_MASTER_STATE:
		f.MasterState(decodeInfo)
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
func (f *FSM) StateDownEvent() { //event *IntfEvent) {
	debug.Logger.Debug(FSM_PREFIX, "handling state down event for:", *f.Config)
	f.StopMasterDownTimer()
	f.StopMasterAdverTimer()

	if f.State == VRRP_MASTER_STATE {
		pkt := f.getPacketInfo()
		pkt.Priority = VRRP_MASTER_DOWN_PRIORITY
		f.SendPkt(pkt)
	}
	f.State = VRRP_INITIALIZE_STATE
	f.DeInitPacketListener()
}

func (f *FSM) StateUpEvent() {
	// during state up event move to initialization
	f.Initialize()
}

func (f *FSM) Exit() {
	f.StateDownEvent()
	f.DeInitPacketListener()
}

func (f *FSM) HandleInterfaceEvent(intfEvent *IntfEvent) {
	switch intfEvent.Event {
	case STATE_CHANGE:
		debug.Logger.Info(FSM_PREFIX, "fsm received state change event", intfEvent.OperState)
		switch intfEvent.OperState {
		case config.STATE_DOWN:
			f.StateDownEvent()
		case config.STATE_UP:
			f.StateUpEvent()
		}
	case CONFIG_CHANGE:
		debug.Logger.Info(FSM_PREFIX, "Changing configuration in fsm:", *intfEvent.Config)
		f.Config = intfEvent.Config

	case TEAR_DOWN:
		// special case...
		debug.Logger.Info(FSM_PREFIX, "Tear down fsm for:", f.Config.IntfRef, "vrid:", f.Config.VRID)
		f.Exit()
	}
}

func (f *FSM) UpdateVirtualIP(enable bool) {
	// Set Sub-intf state up and send out garp via linux stack
	f.vipCh <- &config.VirtualIpInfo{
		IntfRef: f.Config.IntfRef,
		IpAddr:  f.Config.VirtualIPAddr,
		MacAddr: f.VirtualRouterMACAddress,
		Enable:  enable,
		Version: f.Config.Version,
	}
}

func (f *FSM) StartFsm() {
	f.running = true
	f.Initialize()
	for {
		select {
		case pktCh, ok := <-f.pktCh:
			if ok {
				// handle received packet
				f.ProcessRcvdPkt(pktCh)
			}
		case decodeInfo, ok := <-f.decodeCh:
			if ok {
				f.HandleDecodedPkt(decodeInfo)
			}
		case intfStateEv, ok := <-f.IntfEventCh:
			if ok {
				f.HandleInterfaceEvent(intfStateEv)
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

// vrrp states
const (
	VRRP_UNINITIALIZE_STATE = iota
	VRRP_INITIALIZE_STATE
	VRRP_BACKUP_STATE
	VRRP_MASTER_STATE
)

const (
	VRRP_MASTER_PRIORITY         = 255
	VRRP_IGNORE_PRIORITY         = 65535
	VRRP_MASTER_DOWN_PRIORITY    = 0
	VRRP_INITIALIZE_STATE_STRING = "Initialize"
	VRRP_BACKUP_STATE_STRING     = "Backup"
	VRRP_MASTER_STATE_STRING     = "Master"
	VRRP_SNAPSHOT_LEN            = 1024
	VRRP_PROMISCOUS_MODE         = false
	VRRP_TIMEOUT                 = 1 // in seconds
	VRRP_BPF_FILTER              = "ip host " + packet.VRRP_GROUP_IP
	VRRP_MAC_MASK                = "ff:ff:ff:ff:ff:ff"

	FSM_PREFIX = "FSM ------> "
)

const (
	_ = iota
	STATE_CHANGE
	CONFIG_CHANGE
	TEAR_DOWN
)
