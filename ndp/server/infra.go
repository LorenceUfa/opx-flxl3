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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"l3/ndp/config"
	"l3/ndp/debug"
	"utils/commonDefs"
)

/*
 * API: will return all system port information
 */
func (svr *NDPServer) GetPorts() {
	debug.Logger.Info("Get Port State List")
	portsInfo, err := svr.SwitchPlugin.GetAllPortState()
	if err != nil {
		debug.Logger.Err("Failed to get all ports from system, ERROR:", err)
		return
	}
	for _, obj := range portsInfo {
		var empty struct{}
		port := config.PortInfo{
			IntfRef:   obj.IntfRef,
			IfIndex:   obj.IfIndex,
			OperState: obj.OperState,
			Name:      obj.Name,
		}
		pObj, err := svr.SwitchPlugin.GetPort(obj.Name)
		if err != nil {
			debug.Logger.Err("Getting mac address for", obj.Name, "failed, error:", err)
		} else {
			port.MacAddr = pObj.MacAddr
			port.Description = pObj.Description
		}
		l2Port := svr.L2Port[port.IfIndex]
		l2Port.Info = port
		l2Port.RX = nil
		debug.Logger.Info("L2 IfIndex:", port.IfIndex, "Information is:", l2Port.Info)
		svr.L2Port[port.IfIndex] = l2Port
		svr.SwitchMacMapEntries[port.MacAddr] = empty
		svr.SwitchMac = port.MacAddr // @HACK.... need better solution
	}

	debug.Logger.Info("Done with Port State list")
	return
}

/*
 * API: will return all system vlan information
 */
func (svr *NDPServer) GetVlans() {
	debug.Logger.Info("Get Vlan Information")

	// Get Vlan State Information
	vlansStateInfo, err := svr.SwitchPlugin.GetAllVlanState()
	if err != nil {
		debug.Logger.Err("Failed to get system vlan information, ERROR:", err)
		return
	}

	// Get Vlan Config Information
	vlansConfigInfo, err := svr.SwitchPlugin.GetAllVlan()
	if err != nil {
		debug.Logger.Err("Failed to get system vlan config information, ERROR:", err)
	}
	// store vlan state information like name, ifIndex, operstate
	for _, vlanState := range vlansStateInfo {
		entry, _ := svr.VlanInfo[vlanState.IfIndex]
		entry.VlanId = vlanState.VlanId
		entry.VlanIfIndex = vlanState.IfIndex
		entry.Name = vlanState.VlanName
		entry.OperState = vlanState.OperState
		for _, vlanconfig := range vlansConfigInfo {
			if entry.VlanId == vlanconfig.VlanId {
				entry.UntagPortsMap = make(map[int32]bool)
				for _, untagintf := range vlanconfig.UntagIfIndexList {
					entry.UntagPortsMap[untagintf] = true
				}
				entry.TagPortsMap = make(map[int32]bool)
				for _, tagIntf := range vlanconfig.IfIndexList {
					entry.TagPortsMap[tagIntf] = true
				}
			}
		}
		svr.VlanInfo[vlanState.IfIndex] = entry
		svr.VlanIfIdxVlanIdMap[vlanState.VlanName] = vlanState.VlanId
	}
	debug.Logger.Info("Done with Vlan List")
	return
}

/*
 * API: will return all system L3 interfaces information
 */
func (svr *NDPServer) GetIPIntf() {
	debug.Logger.Info("Get IPv6 Interface List")
	ipsInfo, err := svr.SwitchPlugin.GetAllIPv6IntfState()
	if err != nil {
		debug.Logger.Err("Failed to get all ipv6 interfaces from system, ERROR:", err)
		return
	}
	for _, obj := range ipsInfo {
		// ndp will not listen on loopback interfaces
		if svr.SwitchPlugin.IsLoopbackType(obj.IfIndex) {
			continue
		}
		ipInfo, exists := svr.L3Port[obj.IfIndex]
		if !exists {
			ipInfo.InitIntf(obj, svr.PktDataCh, svr.NdpConfig)
			ipInfo.SetIfType(svr.GetIfType(obj.IfIndex))
			// cache reverse map from intfref to ifIndex, used mainly during state
			svr.L3IfIntfRefToIfIndex[obj.IntfRef] = obj.IfIndex
		} else {
			ipInfo.UpdateIntf(obj.IpAddr)
		}
		svr.L3Port[ipInfo.IfIndex] = ipInfo
		if !exists {
			svr.ndpL3IntfStateSlice = append(svr.ndpL3IntfStateSlice, ipInfo.IfIndex)
		}
	}
	debug.Logger.Info("Done with IPv6 State list")
	return
}

func (svr *NDPServer) GetIfType(ifIndex int32) int {
	debug.Logger.Info("get ifType for ifIndex:", ifIndex)
	if _, ok := svr.L2Port[ifIndex]; ok {
		debug.Logger.Info("L3 Port is of IfTypePort")
		return commonDefs.IfTypePort
	}

	if _, ok := svr.VlanInfo[ifIndex]; ok {
		debug.Logger.Info("L3 Port is of IfTypeVlan")
		return commonDefs.IfTypeVlan
	}
	debug.Logger.Info("no valid ifIndex found")
	return -1
}

/*  API: will handle IPv6 notifications received from switch/asicd
 *      Msg types
 *	    1) Create:
 *		    Create an entry in the map
 *	    2) Delete:
 *		    delete an entry from the map
 */
func (svr *NDPServer) HandleIPIntfCreateDelete(obj *config.IPIntfNotification) {
	ipInfo, exists := svr.L3Port[obj.IfIndex]
	switch obj.Operation {
	case config.CONFIG_CREATE:
		// Done during Init
		if exists {
			ipInfo.UpdateIntf(obj.IpAddr)
			svr.L3Port[obj.IfIndex] = ipInfo
			return
		}

		ipInfo = Interface{}
		ipInfo.CreateIntf(obj, svr.PktDataCh, svr.NdpConfig)
		ipInfo.SetIfType(svr.GetIfType(obj.IfIndex))
		// cache reverse map from intfref to ifIndex, used mainly during state
		svr.L3IfIntfRefToIfIndex[obj.IntfRef] = obj.IfIndex
		svr.ndpL3IntfStateSlice = append(svr.ndpL3IntfStateSlice, ipInfo.IfIndex)
	case config.CONFIG_DELETE:
		if !exists {
			debug.Logger.Err("Got Delete request for non existing l3 port", obj.IfIndex)
			return
		}
		// stop rx/tx on the deleted interface
		debug.Logger.Info("Delete IP interface received for", ipInfo.IntfRef, "ifIndex:", ipInfo.IfIndex)
		deleteEntries := ipInfo.DeInitIntf()
		if len(deleteEntries) > 0 {
			svr.DeleteNeighborInfo(deleteEntries, obj.IfIndex)
		}
		delete(svr.L3IfIntfRefToIfIndex, obj.IntfRef)
		// @TODO: need to take care for ifTYpe vlan
		//@TODO: need to remove ndp l3 interface from up slice, but that is taken care of by stop rx/tx
	}
	svr.L3Port[ipInfo.IfIndex] = ipInfo
}

/*  API: will handle l2/physical notifications received from switch/asicd
 *	  Update map entry and then call state notification
 *
 */
func (svr *NDPServer) HandlePhyPortStateNotification(msg *config.PortState) {
	debug.Logger.Info("Handling L2 Port State:", msg.IfState, "for ifIndex:", msg.IfIndex)
	svr.updateL2Operstate(msg.IfIndex, msg.IfState)
}

/*  API: will handle Vlan Create/Delete/Update notifications received from switch/asicd
 */
func (svr *NDPServer) HandleVlanNotification(msg *config.VlanNotification) {
	debug.Logger.Info("Handle Vlan Notfication:", msg.Operation, "for vlanId:", msg.VlanId, "vlan:", msg.VlanName,
		"vlanIfIndex:", msg.VlanIfIndex, "tagList:", msg.TagPorts, "unTagList:", msg.UntagPorts)
	vlan, exists := svr.VlanInfo[msg.VlanIfIndex]
	switch msg.Operation {
	case config.CONFIG_CREATE:
		debug.Logger.Info("Received Vlan Create:", *msg)
		svr.VlanIfIdxVlanIdMap[msg.VlanName] = msg.VlanId
		vlan.Name = msg.VlanName
		vlan.VlanId = msg.VlanId
		vlan.VlanIfIndex = msg.VlanIfIndex
		// Store untag port information
		for _, untagIntf := range msg.UntagPorts {
			if vlan.UntagPortsMap == nil {
				vlan.UntagPortsMap = make(map[int32]bool)
			}
			vlan.UntagPortsMap[untagIntf] = true
		}
		// Store untag port information
		for _, tagIntf := range msg.TagPorts {
			if vlan.TagPortsMap == nil {
				vlan.TagPortsMap = make(map[int32]bool)
			}
			vlan.TagPortsMap[tagIntf] = true
		}
		svr.VlanInfo[msg.VlanIfIndex] = vlan
		svr.Dot1QToVlanIfIndex[msg.VlanId] = msg.VlanIfIndex
	case config.CONFIG_DELETE:
		debug.Logger.Info("Received Vlan Delete:", *msg)
		if exists {
			vlan.UntagPortsMap = nil
			vlan.TagPortsMap = nil
			delete(svr.VlanInfo, msg.VlanIfIndex)
		}
		delete(svr.Dot1QToVlanIfIndex, msg.VlanId)
	case config.CONFIG_UPDATE:
		debug.Logger.Info("Received Vlan Update:", *msg)
		for tagIntf, _ := range vlan.TagPortsMap {
			del := true
			for _, msgTagIntf := range msg.TagPorts {
				if msgTagIntf == tagIntf {
					del = false
					break
				}
			}
			if del {
				delete(vlan.TagPortsMap, tagIntf)
			}
		}
		for unTagIntf, _ := range vlan.UntagPortsMap {
			del := true
			for _, msgUnTagIntf := range msg.UntagPorts {
				if msgUnTagIntf == unTagIntf {
					del = false
					break
				}
			}
			if del {
				delete(vlan.UntagPortsMap, unTagIntf)
			}
		}
		// Store untag port information
		for _, untagIntf := range msg.UntagPorts {
			if vlan.UntagPortsMap == nil {
				vlan.UntagPortsMap = make(map[int32]bool)
			}
			vlan.UntagPortsMap[untagIntf] = true
		}
		// Store untag port information
		for _, tagIntf := range msg.TagPorts {
			if vlan.TagPortsMap == nil {
				vlan.TagPortsMap = make(map[int32]bool)
			}
			vlan.TagPortsMap[tagIntf] = true
		}
		svr.VlanInfo[msg.VlanIfIndex] = vlan
		svr.UpdatePhyPortToVlanInfo(msg)
	}
}

/*  API: will handle IPv6 notifications received from switch/asicd
 *      Msg types
 *	    1) Create:
 *		     Start Rx/Tx in this case
 *	    2) Delete:
 *		     Stop Rx/Tx in this case
 */
func (svr *NDPServer) HandleStateNotification(msg *config.IPIntfNotification) {
	debug.Logger.Info("Handling L3 State:", msg.Operation, "for port:", msg.IntfRef, "ifIndex:", msg.IfIndex, "ipAddr:", msg.IpAddr)
	switch msg.Operation {
	case config.STATE_UP:
		debug.Logger.Info("Create pkt handler for port:", msg.IntfRef, "ifIndex:", msg.IfIndex, "IpAddr:", msg.IpAddr)
		svr.StartRxTx(msg.IfIndex)
	case config.STATE_DOWN:
		debug.Logger.Info("Delete pkt handler for port:", msg.IntfRef, "ifIndex:", msg.IfIndex, "IpAddr:", msg.IpAddr)
		svr.StopRxTx(msg.IfIndex, msg.IpAddr)
	}
}

/*
 *    API: helper function to update ifIndex & port information for software. Hardware is already taken care
 *	   off
 *	   NOTE:
 *         Below Scenario will only happen when mac move happens between a physical port.. L3 port remains
 *	   the same and hence do not need to update clients
 */
func (svr *NDPServer) SoftwareUpdateNbrEntry(msg *config.MacMoveNotification) {
	debug.Logger.Info("Received Mac Move Notification for IPV6 entry:", *msg)
	nbrIp := msg.IpAddr
	svr.NeigborEntryLock.Lock()
	defer svr.NeigborEntryLock.Unlock()
	for _, nbrKey := range svr.neighborKey {
		splitString := splitNeighborKey(nbrKey)
		if splitString[1] == nbrIp {
			debug.Logger.Info("Updating Neigbor information for:", nbrIp, splitString[1])
			nbrEntry, exists := svr.NeighborInfo[nbrKey]
			if !exists {
				return
			}
			l2Port, exists := svr.L2Port[msg.IfIndex]
			if exists {
				nbrEntry.Intf = l2Port.Info.Name
				svr.NeighborInfo[nbrKey] = nbrEntry
				return
			}

			l3Port, exists := svr.L3Port[msg.IfIndex]
			if exists {
				nbrEntry.Intf = l3Port.IntfRef
				svr.NeighborInfo[nbrKey] = nbrEntry
				return
			}
			break
		}
	}
}

/*
 *    API: handle action request coming from the user
 */
func (svr *NDPServer) HandleAction(action *config.ActionData) {
	debug.Logger.Debug("Handle Action:", *action)

	switch action.Type {
	case config.DELETE_BY_IFNAME:
		svr.ActionDeleteByIntf(action.IntfRef)

	case config.DELETE_BY_IPADDR:
		svr.ActionDeleteByNbrIp(action.NbrIp)

	case config.REFRESH_BY_IFNAME:
		svr.ActionRefreshByIntf(action.IntfRef)

	case config.REFRESH_BY_IPADDR:
		svr.ActionRefreshByNbrIp(action.NbrIp)
	}
}

/*
 *    API: It will remove any deleted ip port from the up state slice list
 */
func (svr *NDPServer) DeleteL3IntfFromUpState(ifIndex int32) {
	for idx, entry := range svr.ndpUpL3IntfStateSlice {
		if entry == ifIndex {
			//@TODO: need to optimize this
			svr.ndpUpL3IntfStateSlice = append(svr.ndpUpL3IntfStateSlice[:idx],
				svr.ndpUpL3IntfStateSlice[idx+1:]...)
			break
		}
	}
}

/*
 *    API: It will populate correct vlan information which will be used for ipv6 neighbor create
 */
func (svr *NDPServer) PopulateVlanInfo(nbrInfo *config.NeighborConfig, intfRef string) {
	// check if the ifIndex is present in the reverse map..
	vlanId, exists := svr.VlanIfIdxVlanIdMap[intfRef]
	if exists {
		// if the entry exists then use the vlanId from reverse map
		nbrInfo.VlanId = vlanId
	} else {
		// @TODO: move this to plugin specific
		// in this case use system reserved Vlan id which is -1
		nbrInfo.VlanId = config.INTERNAL_VLAN
	}
}

/*
 *    API: send ipv6 neighbor create notification
 */
func (svr *NDPServer) SendIPv6CreateNotification(ipAddr string, ifIndex int32) {
	msgBuf, err := createNotificationMsg(ipAddr, ifIndex)
	if err != nil {
		return
	}

	notification := commonDefs.NdpNotification{
		MsgType: commonDefs.NOTIFY_IPV6_NEIGHBOR_CREATE,
		Msg:     msgBuf,
	}
	debug.Logger.Info("Sending Create notification for ip address:", ipAddr, "and ifIndex:", ifIndex, "to other protocols")
	svr.pushNotification(notification)
}

/*
 *    API: send ipv6 neighbor delete notification
 */
func (svr *NDPServer) SendIPv6DeleteNotification(ipAddr string, ifIndex int32) {
	msgBuf, err := createNotificationMsg(ipAddr, ifIndex)
	if err != nil {
		return
	}

	notification := commonDefs.NdpNotification{
		MsgType: commonDefs.NOTIFY_IPV6_NEIGHBOR_DELETE,
		Msg:     msgBuf,
	}
	debug.Logger.Info("Sending Delete notification for ip address:", ipAddr, "and ifIndex:", ifIndex, "to other protocols")
	svr.pushNotification(notification)
}

/*
 * helper function to create notification msg
 */
func createNotificationMsg(ipAddr string, ifIndex int32) ([]byte, error) {
	msg := commonDefs.Ipv6NeighborNotification{
		IpAddr:  ipAddr,
		IfIndex: ifIndex,
	}
	msgBuf, err := json.Marshal(msg)
	if err != nil {
		debug.Logger.Err("Failed to marshal IPv6 Neighbor Notification message", msg, "error:", err)
		return msgBuf, err
	}

	return msgBuf, nil
}

/*
 * helper function to marshal notification and push it on to the channel
 */
func (svr *NDPServer) pushNotification(notification commonDefs.NdpNotification) {
	notifyBuf, err := json.Marshal(notification)
	if err != nil {
		debug.Logger.Err("Failed to marshal ipv6 notification before pushing it on channel error:", err)
		return
	}
	svr.notifyChan <- notifyBuf
}

/*
 *  Change L2 port state from switch asicd notification
 */
func (svr *NDPServer) updateL2Operstate(ifIndex int32, state string) {
	l2Port, exists := svr.L2Port[ifIndex]
	if !exists {
		debug.Logger.Err("No L2 Port found for ifIndex:", ifIndex, "hence nothing to update on OperState")
		return
	}
	l2Port.Info.OperState = state
	/* HANDLE PORT FLAP SCENARIOS */
	switch state {
	case config.STATE_UP:
		// NO-OP Just change the state
	case config.STATE_DOWN:
		debug.Logger.Info("L2 Port is down and hence deleting pcap handler for port:", l2Port.Info.Name)
		l2Port.deletePcap()
	}
	debug.Logger.Info("During L2 State Notification L2 IfIndex:", ifIndex, "Information is:", l2Port.Info)
	svr.L2Port[ifIndex] = l2Port
}

/*
 * internal api for creating pcap handler for l2 untagged/tagged physical port for RX
 */
func (l2Port *PhyPort) createPortPcap(pktCh chan *RxPktInfo, name string) (err error) {
	if l2Port.RX == nil {
		debug.Logger.Debug("creating l2 rx pcap for", name, l2Port.Info.IfIndex)
		l2Port.RX, err = pcap.OpenLive(name, NDP_PCAP_SNAPSHOTlEN, NDP_PCAP_PROMISCUOUS, NDP_PCAP_TIMEOUT)
		if err != nil {
			debug.Logger.Err("Creating Pcap Handler failed for l2 interface:", name, "Error:", err)
			return err
		}
		err = l2Port.RX.SetBPFFilter(NDP_PCAP_FILTER)
		if err != nil {
			debug.Logger.Err("Creating BPF Filter failed Error", err)
			l2Port.RX = nil
			return err
		}
		debug.Logger.Info("Created l2 Pcap handler for port:", name, "now start receiving NdpPkts")
		go l2Port.L2ReceiveNdpPkts(pktCh)
	}
	return nil
}

/*
 * internal api for creating pcap handler for l2 physical port for RX
 */
func (l2Port *PhyPort) deletePcap() {
	if l2Port.RX != nil {
		l2Port.RX.Close()
		l2Port.RX = nil
	}
}

/*
 * Receive Ndp Packet and push it on the pktCh
 */
func (intf *PhyPort) L2ReceiveNdpPkts(pktCh chan *RxPktInfo) error {
	if intf.RX == nil {
		debug.Logger.Err("pcap handler for port:", intf.Info.Name, "is not valid. ABORT!!!!")
		return errors.New(fmt.Sprintln("pcap handler for port:", intf.Info.Name, "is not valid. ABORT!!!!"))
	}
	src := gopacket.NewPacketSource(intf.RX, layers.LayerTypeEthernet)
	in := src.Packets()
	for {
		select {
		case pkt, ok := <-in:
			if ok {
				pktCh <- &RxPktInfo{pkt, intf.Info.IfIndex}
			} else {
				debug.Logger.Debug("L2 Pcap closed as in is invalid exiting go routine for port:", intf.Info.Name)
				return nil
			}
		}
	}
	return nil
}

/*
 *  Creating Pcap handlers for l2 port which are marked as tag/untag for l3 vlan port and are in UP state
 *  only l3 CreatePcap should update l2Port.L3 information
 */
func (svr *NDPServer) CreatePcap(ifIndex int32) error {
	debug.Logger.Info("Creating Physical Port Pcap to L3 Vlan, ifIndex:", ifIndex)
	vlan, exists := svr.VlanInfo[ifIndex]
	if !exists {
		debug.Logger.Err("No matching vlan found for ifIndex:", ifIndex)
		return errors.New(fmt.Sprintln("No matching vlan found for ifIndex:", ifIndex))
	}
	debug.Logger.Debug("Creating Pcap Handlers for tag ports:", vlan.TagPortsMap)
	// open rx pcap handler for tagged ports
	for pIfIndex, _ := range vlan.TagPortsMap {
		l2Port, exists := svr.L2Port[pIfIndex]
		if exists {
			debug.Logger.Debug("L2:", l2Port.Info.Name, "operstate is", l2Port.Info.OperState)
			if l2Port.Info.OperState == config.STATE_UP {
				l2Port.createPortPcap(svr.RxPktCh, l2Port.Info.Name)
				svr.L2Port[pIfIndex] = l2Port
			}
		}
	}
	debug.Logger.Debug("Creating pcap handlers for unTag ports:", vlan.UntagPortsMap)
	// open rx pcap handler for untagged ports
	for pIfIndex, _ := range vlan.UntagPortsMap {
		l2Port, exists := svr.L2Port[pIfIndex]
		if exists {
			debug.Logger.Debug("L2:", l2Port.Info.Name, "operstate is", l2Port.Info.OperState)
			if l2Port.Info.OperState == config.STATE_UP {
				l2Port.createPortPcap(svr.RxPktCh, l2Port.Info.Name)
				svr.L2Port[pIfIndex] = l2Port
			}
			// reverse map updated
			l3Info := L3Info{
				Name:    vlan.Name,
				IfIndex: ifIndex,
			}
			svr.L2Port[pIfIndex] = l2Port
			svr.PhyPortToL3PortMap[pIfIndex] = l3Info
		}
	}
	debug.Logger.Debug("PhyPortToL3PortMap after create vlan is:", svr.PhyPortToL3PortMap)
	return nil
}

/*
 *  Deleting Pcap handlers for l2 port which are marked as tag/untag for l3 vlan port and are in UP state
 *  only l3 CreatePcap should update l2Port.L3 information
 */
func (svr *NDPServer) DeletePcap(ifIndex int32) {
	debug.Logger.Info("Deleting Physical Port Pcap RX Handlers for L3 Vlan, ifIndex:", ifIndex)
	vlan, exists := svr.VlanInfo[ifIndex]
	if !exists {
		debug.Logger.Err("No matching vlan found for ifIndex:", ifIndex)
		return //errors.New(fmt.Sprintln("No matching vlan found for ifIndex:", ifIndex))
	}
	debug.Logger.Debug("Deleting pcap handlers for Tag ports:", vlan.TagPortsMap)
	// open rx pcap handler for tagged ports
	for pIfIndex, _ := range vlan.TagPortsMap {
		l2Port, exists := svr.L2Port[pIfIndex]
		if exists {
			l2Port.deletePcap()
			svr.L2Port[pIfIndex] = l2Port
		}
	}
	// open rx pcap handler for untagged ports
	debug.Logger.Debug("Deleting pcap handlers for unTag ports:", vlan.UntagPortsMap)
	for pIfIndex, _ := range vlan.UntagPortsMap {
		l2Port, exists := svr.L2Port[pIfIndex]
		if exists {
			l2Port.deletePcap()
			svr.L2Port[pIfIndex] = l2Port
			delete(svr.PhyPortToL3PortMap, pIfIndex)
		}
	}
}

func (svr *NDPServer) UpdatePhyPortToVlanInfo(msg *config.VlanNotification) {
	ifIndex := msg.VlanIfIndex
	debug.Logger.Info("Updating Phy port to vlan information, for ifIndex:", msg.VlanIfIndex)
	vlan, exists := svr.VlanInfo[msg.VlanIfIndex]
	if !exists {
		debug.Logger.Err("no matching vlan found for update msg:", *msg)
		return
	}
	debug.Logger.Info("vlan tag port information is:", msg.TagPorts)
	// iterating over slice
	for _, pIfIndex := range msg.UntagPorts {
		debug.Logger.Info("Untag port ifIndex:", pIfIndex)
		l3Info, exists := svr.PhyPortToL3PortMap[pIfIndex]
		if !exists {
			l3Info = L3Info{
				Name:    vlan.Name,
				IfIndex: ifIndex,
				Updated: true,
			}
			debug.Logger.Info("new untag port received for pIfIndex:", pIfIndex, "L3Info:", l3Info)
		} else {
			debug.Logger.Info("existing untag port setting update to true for pIfIndex:", pIfIndex, "L3Info:", l3Info)
			l3Info.Updated = true
		}
		svr.PhyPortToL3PortMap[pIfIndex] = l3Info
	}
	l3Port, l3exists := svr.L3Port[ifIndex]
	// iterating over map
	for pIfIndex, l3Info := range svr.PhyPortToL3PortMap {
		l2Port, exists := svr.L2Port[pIfIndex]
		if l3Info.Updated == false {
			debug.Logger.Debug("before deleteing pcap for:", pIfIndex, "check if it belongs to tag port map")
			if _, ok := vlan.TagPortsMap[pIfIndex]; !ok {
				debug.Logger.Info("pIfIndex:", pIfIndex, "is removed from vlan:", ifIndex, vlan.Name, "as un-tag member",
					"hence deleting its pcap")
				l2Port.deletePcap()
			}
			delete(svr.PhyPortToL3PortMap, pIfIndex)
		} else {
			debug.Logger.Info("pIfIndex:", pIfIndex, "is part of vlan:", ifIndex, vlan.Name, "as un-tag member",
				"check if l3 is up or not and start rx/tx if needed")
			l3Info.Updated = false
			svr.PhyPortToL3PortMap[pIfIndex] = l3Info
			// if the entry is updated then check for l3 entry
			if exists {
				/*
				 * if l3 entry exists then check whether l3 is running or not
				 * if l3 is started then incoming should be create
				 * if not then delete the pcap for incoming notification
				 */
				if l3exists {
					if l2Port.Info.OperState == config.STATE_UP && l3Port.PcapBase.Tx != nil {
						l2Port.createPortPcap(svr.RxPktCh, l2Port.Info.Name)
					} else {
						l2Port.deletePcap()
					}
				}
			}
		}
		svr.L2Port[pIfIndex] = l2Port
	}
	// open rx pcap handler for tagged ports
	for pIfIndex, _ := range vlan.TagPortsMap {
		l2Port, exists := svr.L2Port[pIfIndex]
		if exists {
			debug.Logger.Debug("L2:", l2Port.Info.Name, "operstate is", l2Port.Info.OperState)
			if l2Port.Info.OperState == config.STATE_UP && l3Port.PcapBase.Tx != nil {
				l2Port.createPortPcap(svr.RxPktCh, l2Port.Info.Name)
			} else {
				l2Port.deletePcap()
			}
		}
		svr.L2Port[pIfIndex] = l2Port
	}
}

/*
 *  Utility Action function to delete ndp entries by L3 Port interface name
 */
func (svr *NDPServer) ActionDeleteByIntf(intfRef string) {
	ifIndex, exists := svr.L3IfIntfRefToIfIndex[intfRef]
	if !exists {
		debug.Logger.Err("Refresh Action by Interface Name:", intfRef,
			"cannot be performed as no ifIndex found for L3 interface")
		return
	}
	l3Port, exists := svr.L3Port[ifIndex]
	if !exists {
		debug.Logger.Err("Delete Action by Interface Name:", intfRef,
			"cannot be performed as no such L3 interface exists")
		return
	}
	deleteEntries, err := l3Port.FlushNeighbors()
	if len(deleteEntries) > 0 && err == nil {
		debug.Logger.Info("Server Action Delete by Intf:", l3Port.IntfRef, "Neighbors:", deleteEntries)
		svr.DeleteNeighborInfo(deleteEntries, ifIndex)
	}
	svr.L3Port[ifIndex] = l3Port
}

/*
 *  Utility Action function to refreshndp entries by L3 Port interface name
 */
func (svr *NDPServer) ActionRefreshByIntf(intfRef string) {
	ifIndex, exists := svr.L3IfIntfRefToIfIndex[intfRef]
	if !exists {
		debug.Logger.Err("Refresh Action by Interface Name:", intfRef,
			"cannot be performed as no ifIndex found for L3 interface")
		return
	}
	l3Port, exists := svr.L3Port[ifIndex]
	if !exists {
		debug.Logger.Err("Refresh Action by Interface Name:", intfRef,
			"cannot be performed as no such L3 interface exists")
		return
	}

	l3Port.RefreshAllNeighbors(svr.SwitchMac)
	svr.L3Port[ifIndex] = l3Port
}

/*
 *  Utility Action function to delete ndp entries by Neighbor Ip Address
 */
func (svr *NDPServer) ActionDeleteByNbrIp(ipAddr string) {
	var nbrKey string
	found := false
	for _, nbrKey = range svr.neighborKey {
		splitString := splitNeighborKey(nbrKey)
		if splitString[1] == ipAddr {
			found = true
		}
	}
	if !found {
		debug.Logger.Err("Delete Action by Ip Address:", ipAddr, "as no such neighbor is learned")
		return
	}
	nbrEntry, exists := svr.NeighborInfo[nbrKey]
	if !exists {
		debug.Logger.Err("Delete Action by Ip Address:", ipAddr, "as no such neighbor is learned")
		return
	}
	l3IfIndex := nbrEntry.IfIndex
	// if valid vlan then get l3 ifIndex from PhyPortToL3PortMap
	if nbrEntry.VlanId != config.INTERNAL_VLAN {
		l3Info, exists := svr.PhyPortToL3PortMap[nbrEntry.IfIndex]
		if !exists {
			debug.Logger.Err("Delete Action by Ip Address:", ipAddr,
				"cannot be performed as no l3IfIndex mapping found for", nbrEntry.IfIndex,
				"vlan:", nbrEntry.VlanId)
			return
		}
		l3IfIndex = l3Info.IfIndex
	}

	l3Port, exists := svr.L3Port[l3IfIndex]
	if !exists {
		debug.Logger.Err("Delete Action by Ip Address:", ipAddr, "as no L3 Port found where this neighbor is learned")
		return
	}
	deleteEntries, err := l3Port.DeleteNeighbor(nbrEntry)
	if err == nil {
		debug.Logger.Info("Server Action Delete by NbrIp:", ipAddr, "L3 Port:", l3Port.IntfRef,
			"Neighbors:", deleteEntries)
		svr.deleteNeighbor(deleteEntries[0], l3Port.IfIndex)
	}

	svr.L3Port[l3IfIndex] = l3Port
}

/*
 *  Utility Action function to refresh ndp entries by Neighbor Ip Address
 */
func (svr *NDPServer) ActionRefreshByNbrIp(ipAddr string) {
	var nbrKey string
	found := false
	for _, nbrKey = range svr.neighborKey {
		splitString := splitNeighborKey(nbrKey)
		if splitString[1] == ipAddr {
			found = true
		}
	}
	if !found {
		debug.Logger.Err("Delete Action by Ip Address:", ipAddr, "as no such neighbor is learned")
		return
	}
	nbrEntry, exists := svr.NeighborInfo[nbrKey]
	if !exists {
		debug.Logger.Err("Refresh Action by Ip Address:", ipAddr, "as no such neighbor is learned")
		return
	}
	l3IfIndex := nbrEntry.IfIndex
	// if valid vlan then get l3 ifIndex from PhyPortToL3PortMap
	if nbrEntry.VlanId != config.INTERNAL_VLAN {
		l3Info, exists := svr.PhyPortToL3PortMap[nbrEntry.IfIndex]
		if !exists {
			debug.Logger.Err("Refresh Action by Ip Address:", ipAddr,
				"cannot be performed as no l3IfIndex mapping found for", nbrEntry.IfIndex,
				"vlan:", nbrEntry.VlanId)
			return
		}
		l3IfIndex = l3Info.IfIndex
	}

	l3Port, exists := svr.L3Port[l3IfIndex]
	if !exists {
		debug.Logger.Err("Delete Action by Ip Address:", ipAddr, "as no L3 Port found where this neighbor is learned")
		return
	}
	l3Port.SendNS(svr.SwitchMac, nbrEntry.MacAddr, nbrEntry.IpAddr, false /*isFastProbe*/)
	svr.L3Port[l3IfIndex] = l3Port
}
