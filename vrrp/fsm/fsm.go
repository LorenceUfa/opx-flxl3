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
	VRRP_MASTER_PRIORITY         = 255
	VRRP_IGNORE_PRIORITY         = 65535
	VRRP_MASTER_DOWN_PRIORITY    = 0
	VRRP_INITIALIZE_STATE_STRING = "Initialize"
	VRRP_BACKUP_STATE_STRING     = "Backup"
	VRRP_MASTER_STATE_STRING     = "Master"

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
}

type FsmStateInfo struct {
	PktInfo *packet.PacketInfo
}

type IntfEvent struct {
	OperState string
}

type FSM struct {
	Config                  *config.IntfCfg // config attributes like virtual rtr ip, vrid, intfRef, etc...
	pHandle                 *pcap.Handle    // Pcap Handler for receiving packets
	IpAddr                  string          // My own ip address
	IfIndex                 int32           // My own ifIndex
	VirtualRouterMACAddress string          // VRRP MAC aka VMAC
	State                   uint8
	PktInfo                 *PacketInfo
	AdverTimer              *time.Timer // Advertisement Timer
	MasterAdverInterval     int32       // The initial value is the same as Advertisement_Interval.
	SkewTime                int32       // (((256 - priority) * Master_Adver_Interval) / 256)
	MasterDownValue         int32       // (3 * Master_Adver_Interval) + Skew_time
	MasterDownTimer         *time.Timer
	StInfo                  *config.State   // this is state information for this fsm which will be used for get bulk
	StateCh                 chan *IntfState // push current state information on to this channel so that server can update information
	pktCh                   *PktChannelInfo
	fsmStCh                 *FsmStateInfo
	txPktCh                 *packet.PacketInfo
	IntfEventCh             *IntfEvent
}

func InitFsm(cfg *config.IntfCfg, l3Info *L3Intf, stCh chan *IntfState) *FSM {
	f := &FSM{}
	f.Config = *cfg
	f.StInfo = &config.State{}
	f.IpAddr = l3Info.IpAddr
	f.IfIndex = l3Info.IfIndex
	f.VirtualRouterMACAddress = createVirtualMac(cfg.VRID)
	f.StateCh = stCh
	f.pktCh = make(chan *PktChannelInfo)
	f.fsmStCh = make(chan *FsmStateInfo)
	f.txPktCh = make(chan *packet.PacketInfo)
	f.IntfEventCh = make(chan *IntfEvent)
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

func (f *FSM) UpdateRxStateInformation(pktInfo *packet.PacketInfo) {
	f.StInfo.MasterIp = pktInfo.IpAddr
	f.StInfo.AdverRx++
	f.StInfo.LastAdverRx = time.Now().String()
	f.StInfo.CurrentFsmState = getStateName(f.State)
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

func (f *FSM) CalculateDownValue(advInt int32) {
	//(155) + Set Master_Adver_Interval to Advertisement_Interval
	if advInt == config.USE_CONFIG_ADVERTISEMENT {
		f.MasterAdverInterval = f.Config.AdvertisementInterval
	} else {
		f.MasterAdverInterval = advInt
	}
	//(160) + Set the Master_Down_Timer to Master_Down_Interval
	if f.Config.Priority != 0 && f.Config.MasterAdverInterval != 0 {
		f.Config.SkewTime = ((256 - f.Config.Priority) * f.Config.MasterAdverInterval) / 256
	}
	f.MasterDownValue = (3 * f.MasterAdverInterval) + f.SkewTime
}

func (f *FSM) Initialize() {
	debug.Logger.Debug(FSM_PREFIX, "In Init state deciding next state")
	switch f.Config.Priority {
	case VRRP_MASTER_PRIORITY:
		f.TransitionToMaster()
	default:
		f.TransitionToBackup(config.USE_CONFIG_ADVERTISEMENT)
	}
}

func (f *FSM) ProcessStateInfo(fsmStInfo *FsmStateInfo) {
	debug.Logger.Debug(FSM_PREFIX, "Processing State Information")
	switch f.State {
	case VRRP_INITIALIZE_STATE:
		f.Initialize()
	case VRRP_BACKUP_STATE:
		f.BackupState(fsmStInfo)
	case VRRP_MASTER_STATE:
		f.MasterState(fsmStInfo)
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
func (f *FSM) StateDownEvent(event *IntfEvent) {
	f.StopMasterDownTimer()
	f.StopMasterAdverTimer()

	if f.State == VRRP_MASTER_STATE {
		pkt := f.getPacketInfo()
		pkt.Priority = VRRP_MASTER_DOWN_PRIORITY
		f.SendPkt(pkt)
	}
	f.State = VRRP_INITIALIZE_STATE
}

func (f *FSM) StateUpEvent(event *IntfEvent) {
	// during state up event move to initialization
	f.Initialize()
}

func (f *FSM) HandleInterfaceEvent(event *IntfEvent) {
	switch event.OperState {
	case config.STATE_DOWN:
		f.StateDownEvent(event)
	case config.STATE_UP:
	}
}

func (f *FSM) StartFsm() {
	for {
		select {
		case pktCh, ok := <-f.pktCh:
			if ok {
				// handle received packet
				f.ProcessRcvdPkt(pktCh)
			}
		case fsmStInfo, ok := <-f.fsmStCh:
			if ok {
				f.ProcessStateInfo(fsmStInfo)
			}

		case intfStateEv, ok := <-f.IntfEventCh:
			if ok {
				f.HandleInterfaceEvent(intfStateEv)
			}
		}
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
