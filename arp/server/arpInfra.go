//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//       Unless required by applicable law or agreed to in writing, software
//       distributed under the License is distributed on an "AS IS" BASIS,
//       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//       See the License for the specific language governing permissions and
//       limitations under the License.
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
	"github.com/google/gopacket/pcap"
	"net"
	"utils/commonDefs"
)

const (
	UNTAGGED bool = false
	TAGGED   bool = true
)

type L3VlanStruct struct {
	IpAddr  string
	IpNet   net.IPMask
	L3IfIdx int
}

type L3LagStruct struct {
	L3Vlan   L3VlanStruct
	LagIfIdx int
	VlanId   int
	TagFlag  bool
}

type L3PortStruct struct {
	L3Lag     L3LagStruct
	PortIfIdx int
}

type L3IntfProperty struct {
	Netmask   net.IPMask
	IpAddr    string
	IfName    string
	OperState bool
}

type L3PortProp struct {
	L3IfIdx  int
	IpAddr   string
	Netmask  net.IPMask
	LagIfIdx int
}

type PortProperty struct {
	IfName        string
	MacAddr       string
	UntagVlanId   int
	L3PortPropMap map[int]L3PortProp //VlanId
	CtrlCh        chan bool
	CtrlReplyCh   chan bool
	PcapHdl       *pcap.Handle
	OperState     bool

	//TO be deleted
	L3IfIdx int
}

type VlanProperty struct {
	IfName        string
	UntagIfIdxMap map[int]bool
	TagIfIdxMap   map[int]bool
}

type LagProperty struct {
	IfName    string
	VlanIdMap map[int]bool
	PortMap   map[int]bool
}

func (server *ARPServer) buildArpInfra() {
	server.constructPortInfra()
	server.constructLagInfra()
	server.constructVlanInfra()
	server.constructL3Infra()
}

func (server *ARPServer) constructPortInfra() {
	server.getBulkPortState()
	server.getBulkPortConfig()
}

func (server *ARPServer) GetMacAddr(l3IfIdx int) string {
	ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
	switch ifType {
	case commonDefs.IfTypeVlan:
		vlanEnt, _ := server.vlanPropMap[l3IfIdx]
		for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
			uIfType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(uIfIdx))
			switch uIfType {
			case commonDefs.IfTypeLag:
				lagEnt, _ := server.lagPropMap[uIfIdx]
				for portIfIdx, _ := range lagEnt.PortMap {
					portEnt, _ := server.portPropMap[portIfIdx]
					return portEnt.MacAddr
				}
			case commonDefs.IfTypePort:
				portEnt, _ := server.portPropMap[uIfIdx]
				return portEnt.MacAddr
			}
		}
		for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
			tIfType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(tIfIdx))
			switch tIfType {
			case commonDefs.IfTypeLag:
				lagEnt, _ := server.lagPropMap[tIfIdx]
				for portIfIdx, _ := range lagEnt.PortMap {
					portEnt, _ := server.portPropMap[portIfIdx]
					return portEnt.MacAddr
				}
			case commonDefs.IfTypePort:
				portEnt, _ := server.portPropMap[tIfIdx]
				return portEnt.MacAddr
			}
		}
	case commonDefs.IfTypeLag:
		lagEnt, _ := server.lagPropMap[l3IfIdx]
		for portIfIdx, _ := range lagEnt.PortMap {
			portEnt, _ := server.portPropMap[portIfIdx]
			return portEnt.MacAddr
		}
	case commonDefs.IfTypePort:
		portEnt, _ := server.portPropMap[l3IfIdx]
		return portEnt.MacAddr
	}
	return ""
}

func (server *ARPServer) getBulkPortState() {
	curMark := int(asicdCommonDefs.MIN_SYS_PORTS)
	server.logger.Debug("Calling Asicd for getting Port Property")
	count := 100
	for {
		bulkInfo, _ := server.AsicdPlugin.GetBulkPortState(curMark, count)
		if bulkInfo == nil {
			break
		}
		objCnt := int(bulkInfo.Count)
		more := bool(bulkInfo.More)
		curMark = int(bulkInfo.EndIdx)
		for i := 0; i < objCnt; i++ {
			ifIndex := int(bulkInfo.PortStateList[i].IfIndex)
			ent := server.portPropMap[ifIndex]
			ent.IfName = bulkInfo.PortStateList[i].Name
			ent.UntagVlanId = -1
			ent.L3PortPropMap = make(map[int]L3PortProp)
			ent.PcapHdl = nil
			switch bulkInfo.PortStateList[i].OperState {
			case "UP":
				ent.OperState = true
			case "DOWN":
				ent.OperState = false
			default:
				server.logger.Err("Invalid OperState for the port",
					bulkInfo.PortStateList[i].OperState, ent.IfName)
			}
			server.portPropMap[ifIndex] = ent
		}
		if more == false {
			break
		}
	}
}

func (server *ARPServer) getBulkPortConfig() {
	curMark := int(asicdCommonDefs.MIN_SYS_PORTS)
	server.logger.Debug("Calling Asicd for getting Port Property")
	count := 100
	for {
		bulkInfo, _ := server.AsicdPlugin.GetBulkPort(curMark, count)
		if bulkInfo == nil {
			break
		}
		objCnt := int(bulkInfo.Count)
		more := bool(bulkInfo.More)
		curMark = int(bulkInfo.EndIdx)
		for i := 0; i < objCnt; i++ {
			ifIndex := int(bulkInfo.PortList[i].IfIndex)
			ent := server.portPropMap[ifIndex]
			ent.MacAddr = bulkInfo.PortList[i].MacAddr
			ent.CtrlCh = make(chan bool)
			ent.CtrlReplyCh = make(chan bool)
			server.portPropMap[ifIndex] = ent
		}
		if more == false {
			break
		}
	}
}

func (server *ARPServer) constructLagInfra() {
	curMark := 0
	server.logger.Info("Calling Asicd for getting Lag Property")
	count := 100
	for {
		bulkLagInfo, _ := server.AsicdPlugin.GetBulkLag(curMark, count)
		if bulkLagInfo == nil {
			break
		}
		objCnt := int(bulkLagInfo.Count)
		more := bool(bulkLagInfo.More)
		curMark = int(bulkLagInfo.EndIdx)
		for i := 0; i < objCnt; i++ {
			lagIfIdx := int(bulkLagInfo.LagList[i].LagIfIndex)
			lagEnt := server.lagPropMap[lagIfIdx]
			lagEnt.IfName = bulkLagInfo.LagList[i].LagName
			lagEnt.VlanIdMap = make(map[int]bool)
			ifIndexList := bulkLagInfo.LagList[i].IfIndexList
			for i := 0; i < len(ifIndexList); i++ {
				ifIdx := int(ifIndexList[i])
				lagEnt.PortMap[ifIdx] = true
			}
			server.lagPropMap[lagIfIdx] = lagEnt
		}
		if more == false {
			break
		}
	}
}

func (server *ARPServer) constructVlanInfra() {
	curMark := 0
	server.logger.Debug("Calling Asicd for getting Vlan Property")
	count := 100
	for {
		bulkVlanInfo, _ := server.AsicdPlugin.GetBulkVlan(curMark, count)
		if bulkVlanInfo == nil {
			break
		}
		/*
		* Getbulk on vlan state assuming indexes are one to on mapped
		* between state and config object
		 */
		bulkVlanStateInfo, _ := server.AsicdPlugin.GetBulkVlanState(curMark, count)
		if bulkVlanStateInfo == nil {
			break
		}
		objCnt := int(bulkVlanInfo.Count)
		more := bool(bulkVlanInfo.More)
		curMark = int(bulkVlanInfo.EndIdx)
		for i := 0; i < objCnt; i++ {
			vlanIfIdx := int(bulkVlanStateInfo.VlanStateList[i].IfIndex)
			vlanEnt := server.vlanPropMap[vlanIfIdx]
			vlanEnt.IfName = bulkVlanStateInfo.VlanStateList[i].VlanName
			untaggedIfIndexList := bulkVlanInfo.VlanList[i].UntagIfIndexList
			vlanEnt.UntagIfIdxMap = make(map[int]bool)
			for i := 0; i < len(untaggedIfIndexList); i++ {
				vlanEnt.UntagIfIdxMap[int(untaggedIfIndexList[i])] = true
			}
			taggedIfIndexList := bulkVlanInfo.VlanList[i].IfIndexList
			vlanEnt.TagIfIdxMap = make(map[int]bool)
			for i := 0; i < len(taggedIfIndexList); i++ {
				vlanEnt.TagIfIdxMap[int(taggedIfIndexList[i])] = true
			}
			server.vlanPropMap[vlanIfIdx] = vlanEnt
		}
		if more == false {
			break
		}
	}
}

func (server *ARPServer) constructL3Infra() {
	currMark := 0
	server.logger.Debug("Calling Asicd for getting L3 Interfaces")
	count := 100
	for {
		bulkIPv4Info, _ := server.AsicdPlugin.GetBulkIPv4IntfState(currMark, count)
		if bulkIPv4Info == nil {
			break
		}

		objCnt := int(bulkIPv4Info.Count)
		more := bool(bulkIPv4Info.More)
		currMark = int(bulkIPv4Info.EndIdx)
		for i := 0; i < objCnt; i++ {
			server.processIPv4IntfCreate(bulkIPv4Info.IPv4IntfStateList[i].IpAddr, bulkIPv4Info.IPv4IntfStateList[i].IfIndex)
			l3IfIdx := int(bulkIPv4Info.IPv4IntfStateList[i].IfIndex)
			l3Ent := server.l3IntfPropMap[l3IfIdx]
			switch bulkIPv4Info.IPv4IntfStateList[i].OperState {
			case "UP":
				l3Ent.OperState = true
			case "DOWN":
				l3Ent.OperState = false
			default:
				server.logger.Err("Invalid OperState for the L3 Interface")
			}
			server.l3IntfPropMap[l3IfIdx] = l3Ent
		}
		if more == false {
			break
		}
	}
}

func (server *ARPServer) updateVlanWithL3(l3Vlan L3VlanStruct) string {
	vlanEnt := server.vlanPropMap[l3Vlan.L3IfIdx]
	vlanId := asicdCommonDefs.GetIntfIdFromIfIndex(int32(l3Vlan.L3IfIdx))
	l3Lag := L3LagStruct{
		L3Vlan: l3Vlan,
		VlanId: vlanId,
	}
	for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
		l3Lag.TagFlag = UNTAGGED
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(uIfIdx))
		switch ifType {
		case commonDefs.IfTypeLag:
			l3Lag.LagIfIdx = uIfIdx
			server.updateLagWithL3(l3Lag)
		case commonDefs.IfTypePort:
			l3Lag.LagIfIdx = -1
			l3Port := L3PortStruct{
				L3Lag:     l3Lag,
				PortIfIdx: uIfIdx,
			}
			server.updatePortWithL3(l3Port)
		}
	}
	for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
		l3Lag.TagFlag = TAGGED
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(tIfIdx))
		switch ifType {
		case commonDefs.IfTypeLag:
			l3Lag.LagIfIdx = tIfIdx
			server.updateLagWithL3(l3Lag)
		case commonDefs.IfTypePort:
			l3Lag.LagIfIdx = -1
			l3Port := L3PortStruct{
				L3Lag:     l3Lag,
				PortIfIdx: tIfIdx,
			}
			server.updatePortWithL3(l3Port)
		}
	}
	return vlanEnt.IfName
}

func (server *ARPServer) updateLagWithL3(l3Lag L3LagStruct) string {
	lagEnt := server.lagPropMap[l3Lag.LagIfIdx]
	l3Port := L3PortStruct{
		L3Lag: l3Lag,
	}
	lagEnt.VlanIdMap[l3Lag.VlanId] = true
	for portIfIdx, _ := range lagEnt.PortMap {
		l3Port.PortIfIdx = portIfIdx
		server.updatePortWithL3(l3Port)
	}
	return lagEnt.IfName
}

func (server *ARPServer) updatePortWithL3(l3Port L3PortStruct) string {
	portEnt := server.portPropMap[l3Port.PortIfIdx]
	ifName := portEnt.IfName
	l3PortPropEnt := portEnt.L3PortPropMap[l3Port.L3Lag.VlanId]
	l3PortPropEnt.IpAddr = l3Port.L3Lag.L3Vlan.IpAddr
	l3PortPropEnt.Netmask = l3Port.L3Lag.L3Vlan.IpNet
	l3PortPropEnt.LagIfIdx = l3Port.L3Lag.LagIfIdx
	l3PortPropEnt.L3IfIdx = l3Port.L3Lag.L3Vlan.L3IfIdx
	portEnt.L3PortPropMap[l3Port.L3Lag.VlanId] = l3PortPropEnt
	if l3Port.L3Lag.TagFlag == UNTAGGED {
		portEnt.UntagVlanId = l3Port.L3Lag.VlanId
	}
	server.portPropMap[l3Port.PortIfIdx] = portEnt
	return ifName
}

func (server *ARPServer) processIPv4NbrMacMove(msg commonDefs.IPv4NbrMacMoveNotifyMsg) {
	server.arpEntryMacMoveCh <- msg
}

func (server *ARPServer) processArpInfra() {
	for l3IfIdx, l3Ent := range server.l3IntfPropMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if l3Ent.OperState == true {
			server.EnableArpOnL3(l3IfIdx, ifType)
		}
	}
}

func (server *ARPServer) updateIPv4Infra(msg commonDefs.IPv4IntfNotifyMsg) {
	if msg.MsgType == commonDefs.NOTIFY_IPV4INTF_CREATE {
		server.processIPv4IntfCreate(msg.IpAddr, msg.IfIndex)
	} else {
		server.processIPv4IntfDelete(msg.IpAddr, msg.IfIndex)
	}
}

func (server *ARPServer) processIPv4IntfCreate(IpAddr string, IfIndex int32) {
	var ifName string
	ip, ipNet, _ := net.ParseCIDR(IpAddr)
	l3IfIdx := int(IfIndex)
	ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(IfIndex)
	ipAddr := ip.String()
	l3Vlan := L3VlanStruct{
		IpAddr:  ipAddr,
		IpNet:   ipNet.Mask,
		L3IfIdx: l3IfIdx,
	}
	switch ifType {
	case commonDefs.IfTypeVlan:
		ifName = server.updateVlanWithL3(l3Vlan)
	case commonDefs.IfTypeLag:
		l3Lag := L3LagStruct{
			L3Vlan:   l3Vlan,
			LagIfIdx: l3IfIdx,
			VlanId:   -1,
			TagFlag:  UNTAGGED,
		}
		ifName = server.updateLagWithL3(l3Lag)
	case commonDefs.IfTypePort:
		l3Lag := L3LagStruct{
			L3Vlan:   l3Vlan,
			LagIfIdx: -1,
			VlanId:   -1,
			TagFlag:  UNTAGGED,
		}
		l3Port := L3PortStruct{
			L3Lag:     l3Lag,
			PortIfIdx: l3IfIdx,
		}
		ifName = server.updatePortWithL3(l3Port)
	}
	l3Ent := server.l3IntfPropMap[l3IfIdx]
	l3Ent.Netmask = ipNet.Mask
	l3Ent.IpAddr = ipAddr
	l3Ent.IfName = ifName
	server.l3IntfPropMap[l3IfIdx] = l3Ent
}

func (server *ARPServer) processIPv4IntfDelete(IpAddr string, IfIndex int32) {
	l3IfIdx := int(IfIndex)
	ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(IfIndex)
	switch ifType {
	case commonDefs.IfTypeVlan:
		server.deleteVlanWithL3(l3IfIdx)
	case commonDefs.IfTypeLag:
		server.deleteLagWithL3(l3IfIdx, l3IfIdx, -1)
	case commonDefs.IfTypePort:
		server.deletePortWithL3(l3IfIdx, l3IfIdx, -1)
	}
	delete(server.l3IntfPropMap, l3IfIdx)
}

func (server *ARPServer) deleteVlanWithL3(l3IfIdx int) {
	vlanEnt := server.vlanPropMap[l3IfIdx]
	vlanId := asicdCommonDefs.GetIntfIdFromIfIndex(int32(l3IfIdx))
	for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if ifType == commonDefs.IfTypeLag {
			server.deleteLagWithL3(l3IfIdx, uIfIdx, vlanId)
		} else {
			//server.deletePortWithL3(l3IfIdx, -1, uIfIdx, vlanId)
			server.deletePortWithL3(l3IfIdx, uIfIdx, vlanId)
		}
	}
	for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if ifType == commonDefs.IfTypeLag {
			server.deleteLagWithL3(l3IfIdx, tIfIdx, vlanId)
		} else {
			//server.deletePortWithL3(l3IfIdx, -1, tIfIdx, vlanId)
			server.deletePortWithL3(l3IfIdx, tIfIdx, vlanId)
		}
	}
}

func (server *ARPServer) deleteLagWithL3(l3IfIdx int, lagIfIdx int, vlanId int) {
	lagEnt := server.lagPropMap[lagIfIdx]
	delete(lagEnt.VlanIdMap, vlanId)
	for portIfIdx, _ := range lagEnt.PortMap {
		server.deletePortWithL3(l3IfIdx, portIfIdx, vlanId)
	}
}

func (server *ARPServer) deletePortWithL3(l3IfIdx int, portIfIdx int, vlanId int) {
	portEnt := server.portPropMap[portIfIdx]
	delete(portEnt.L3PortPropMap, vlanId)
	server.portPropMap[portIfIdx] = portEnt
}

func (server *ARPServer) processIPv4L3StateChange(msg commonDefs.IPv4L3IntfStateNotifyMsg) {
	ifIdx := int(msg.IfIndex)
	ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(msg.IfIndex)
	if msg.IfState == 0 {
		server.DisableArpOnL3(ifIdx, ifType)
		server.DisableL3(ifIdx)
	} else {
		server.EnableL3(ifIdx)
		server.EnableArpOnL3(ifIdx, ifType)
	}
}

func (server *ARPServer) processL2StateChange(msg commonDefs.L2IntfStateNotifyMsg) {
	ifIdx := int(msg.IfIndex)
	ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(msg.IfIndex)
	if msg.IfState == 0 {
		switch ifType {
		case commonDefs.IfTypeVlan:
			server.DisableArpOnVlan(ifIdx)
		case commonDefs.IfTypeLag:
			server.DisableArpOnLag(-1, ifIdx)
		case commonDefs.IfTypePort:
			server.DisableArpOnPort(-1, ifIdx)
			server.DisablePort(ifIdx)
		}
	} else {
		switch ifType {
		case commonDefs.IfTypeVlan:
			server.EnableArpOnVlan(ifIdx)
		case commonDefs.IfTypeLag:
			server.EnableArpOnLag(-1, ifIdx)
		case commonDefs.IfTypePort:
			server.EnablePort(ifIdx)
			server.EnableArpOnPort(-1, ifIdx)
		}

	}
}

func (server *ARPServer) DisableArpOnL3(l3IfIdx int, ifType int) {
	switch ifType {
	case commonDefs.IfTypeVlan:
		server.DisableArpOnVlan(l3IfIdx)
	case commonDefs.IfTypeLag:
		server.DisableArpOnLag(l3IfIdx, l3IfIdx)
	case commonDefs.IfTypePort:
		server.DisableArpOnPort(l3IfIdx, l3IfIdx)
	}
}

func (server *ARPServer) DisableArpOnVlan(l3IfIdx int) {
	vlanEnt := server.vlanPropMap[l3IfIdx]
	for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if ifType == commonDefs.IfTypeLag {
			server.DisableArpOnLag(l3IfIdx, uIfIdx)
		} else {
			server.DisableArpOnPort(l3IfIdx, uIfIdx)
		}
	}
	for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if ifType == commonDefs.IfTypeLag {
			server.DisableArpOnLag(l3IfIdx, tIfIdx)
		} else {
			server.DisableArpOnPort(l3IfIdx, tIfIdx)
		}
	}
}

func (server *ARPServer) DisableArpOnLag(l3IfIdx int, lagIfIdx int) {
	lagEnt := server.lagPropMap[lagIfIdx]
	for portIfIdx, _ := range lagEnt.PortMap {
		server.DisableArpOnPort(l3IfIdx, portIfIdx)
	}
}

func (server *ARPServer) DisableArpOnPort(l3IfIdx int, portIfIdx int) {
	//var err error
	_, exist := server.l3IntfPropMap[l3IfIdx]
	if !exist {
		return
	}
	flag := false
	var vlanId int
	portEnt := server.portPropMap[portIfIdx]
	if portEnt.OperState == true {
		if portEnt.PcapHdl != nil {
			if l3IfIdx != -1 {
				for id, l3PortProp := range portEnt.L3PortPropMap {
					if l3PortProp.L3IfIdx == l3IfIdx {
						vlanId = id
						continue
					}
					l3Ent, _ := server.l3IntfPropMap[l3PortProp.L3IfIdx]
					if l3Ent.OperState == true {
						flag = true
					}
				}
			}
			if flag == false {
				portEnt.CtrlCh <- true
				<-portEnt.CtrlReplyCh
				portEnt.PcapHdl.Close()
				portEnt.PcapHdl = nil
			}
			server.arpEntryDeleteCh <- DeleteArpEntryMsg{
				PortIfIdx: portIfIdx,
				VlanId:    vlanId,
				L3IfIdx:   l3IfIdx,
			}
		}
	}
	server.portPropMap[portIfIdx] = portEnt
}

func (server *ARPServer) EnableArpOnL3(l3IfIdx int, ifType int) {
	switch ifType {
	case commonDefs.IfTypeVlan:
		server.EnableArpOnVlan(l3IfIdx)
	case commonDefs.IfTypeLag:
		server.EnableArpOnLag(l3IfIdx, l3IfIdx)
	case commonDefs.IfTypePort:
		server.EnableArpOnPort(l3IfIdx, l3IfIdx)
	}
}

func (server *ARPServer) EnableArpOnVlan(l3IfIdx int) {
	vlanEnt := server.vlanPropMap[l3IfIdx]
	for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if ifType == commonDefs.IfTypeLag {
			server.EnableArpOnLag(l3IfIdx, uIfIdx)
		} else {
			server.EnableArpOnPort(l3IfIdx, uIfIdx)
		}
	}
	for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
		ifType := asicdCommonDefs.GetIntfTypeFromIfIndex(int32(l3IfIdx))
		if ifType == commonDefs.IfTypeLag {
			server.EnableArpOnLag(l3IfIdx, tIfIdx)
		} else {
			server.EnableArpOnPort(l3IfIdx, tIfIdx)
		}
	}
}

func (server *ARPServer) EnableArpOnLag(l3IfIdx int, lagIfIdx int) {
	lagEnt := server.lagPropMap[lagIfIdx]
	for portIfIdx, _ := range lagEnt.PortMap {
		server.EnableArpOnPort(l3IfIdx, portIfIdx)
	}
}

func (server *ARPServer) EnableArpOnPort(l3IfIdx int, portIfIdx int) {
	var err error
	_, exist := server.l3IntfPropMap[l3IfIdx]
	if !exist {
		return
	}
	portEnt := server.portPropMap[portIfIdx]
	if portEnt.OperState == true {
		if portEnt.PcapHdl == nil {
			portEnt.PcapHdl, err = server.StartArpRxTx(portEnt.IfName, portEnt.MacAddr)
			if err != nil {
				server.logger.Err("Error opening pcap handle on", portEnt.IfName, err)
				return
			}
			server.portPropMap[portIfIdx] = portEnt
		}
		go server.processRxPkts(portIfIdx)
		go server.SendArpProbe(l3IfIdx, portEnt.MacAddr)
	}
}

func (server *ARPServer) DisablePort(portIfIdx int) {
	portEnt := server.portPropMap[portIfIdx]
	portEnt.OperState = false
	server.portPropMap[portIfIdx] = portEnt
}

func (server *ARPServer) EnablePort(portIfIdx int) {
	portEnt := server.portPropMap[portIfIdx]
	portEnt.OperState = true
	server.portPropMap[portIfIdx] = portEnt
}

func (server *ARPServer) EnableL3(l3IfIdx int) {
	l3Ent, _ := server.l3IntfPropMap[l3IfIdx]
	l3Ent.OperState = true
	server.l3IntfPropMap[l3IfIdx] = l3Ent
}

func (server *ARPServer) DisableL3(l3IfIdx int) {
	l3Ent, _ := server.l3IntfPropMap[l3IfIdx]
	l3Ent.OperState = false
	server.l3IntfPropMap[l3IfIdx] = l3Ent
}

func (server *ARPServer) updateVlanInfra(msg commonDefs.VlanNotifyMsg) {
	vlanId := int(msg.VlanId)
	vlanIfIdx := int(asicdCommonDefs.GetIfIndexFromIntfIdAndIntfType(vlanId, commonDefs.IfTypeVlan))
	uIfIdxList := msg.UntagPorts
	tIfIdxList := msg.TagPorts
	vlanName := msg.VlanName
	if msg.MsgType == commonDefs.NOTIFY_VLAN_CREATE {
		server.processVlanCreate(vlanName, vlanIfIdx, uIfIdxList, tIfIdxList)
	} else if msg.MsgType == commonDefs.NOTIFY_VLAN_UPDATE {
		server.processVlanUpdate(vlanName, vlanIfIdx, uIfIdxList, tIfIdxList)
	} else if msg.MsgType == commonDefs.NOTIFY_VLAN_DELETE {
		server.processVlanDelete(vlanName, vlanIfIdx, uIfIdxList, tIfIdxList)
	}
}

func (server *ARPServer) updateLagInfra(msg commonDefs.LagNotifyMsg) {
	lagIfIdx := int(msg.IfIndex)
	portList := msg.IfIndexList
	if msg.MsgType == commonDefs.NOTIFY_LAG_CREATE {
		server.processLagCreate(lagIfIdx, portList)
	} else if msg.MsgType == commonDefs.NOTIFY_LAG_UPDATE {
		server.processLagUpdate(lagIfIdx, portList)
	} else if msg.MsgType == commonDefs.NOTIFY_LAG_DELETE {
		server.processLagDelete(lagIfIdx, portList)
	}
}

func (server *ARPServer) processVlanCreate(vlanName string, vlanIfIdx int, uIfIdxList, tIfIdxList []int32) {
	vlanEnt, _ := server.vlanPropMap[vlanIfIdx]
	vlanEnt.UntagIfIdxMap = nil
	vlanEnt.UntagIfIdxMap = make(map[int]bool)
	for idx := 0; idx < len(uIfIdxList); idx++ {
		vlanEnt.UntagIfIdxMap[int(uIfIdxList[idx])] = true
	}
	vlanEnt.TagIfIdxMap = nil
	vlanEnt.TagIfIdxMap = make(map[int]bool)
	for idx := 0; idx < len(tIfIdxList); idx++ {
		vlanEnt.TagIfIdxMap[int(tIfIdxList[idx])] = true
	}
	vlanEnt.IfName = vlanName
	server.vlanPropMap[vlanIfIdx] = vlanEnt
	l3IntfFlag, l3IntfOperState := server.IsL3Intf(vlanIfIdx)
	if l3IntfFlag == true && l3IntfOperState == true {
		l3Ent, _ := server.l3IntfPropMap[vlanIfIdx]
		l3Vlan := L3VlanStruct{
			IpAddr:  l3Ent.IpAddr,
			IpNet:   l3Ent.Netmask,
			L3IfIdx: vlanIfIdx,
		}
		server.updateVlanWithL3(l3Vlan)
		server.EnableArpOnVlan(vlanIfIdx)
	}
}

func (server *ARPServer) IsL3Intf(ifIdx int) (bool, bool) {
	l3Ent, exist := server.l3IntfPropMap[ifIdx]
	if !exist {
		return false, l3Ent.OperState
	}
	return true, l3Ent.OperState
}

func (server *ARPServer) processVlanDelete(vlanName string, vlanIfIdx int, uIfIdxList, tIfIdxList []int32) {
	vlanEnt, _ := server.vlanPropMap[vlanIfIdx]
	l3IntfFlag, l3IntfOperState := server.IsL3Intf(vlanIfIdx)
	if l3IntfFlag {
		if l3IntfOperState {
			server.DisableArpOnVlan(vlanIfIdx)
		}
		server.deleteVlanWithL3(vlanIfIdx)
	}
	for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
		delete(vlanEnt.UntagIfIdxMap, uIfIdx)
	}
	vlanEnt.UntagIfIdxMap = nil
	for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
		delete(vlanEnt.TagIfIdxMap, tIfIdx)
	}
	vlanEnt.TagIfIdxMap = nil
	delete(server.vlanPropMap, vlanIfIdx)
}

func (server *ARPServer) processVlanUpdate(vlanName string, vlanIfIdx int, uIfIdxList, tIfIdxList []int32) {
	var uIfIdxDelList []int32
	var tIfIdxDelList []int32
	var uIfIdxCreateList []int32
	var tIfIdxCreateList []int32
	newUIfIdxMap := make(map[int]bool)
	for idx := 0; idx < len(uIfIdxList); idx++ {
		newUIfIdxMap[int(uIfIdxList[idx])] = true
	}
	newTIfIdxMap := make(map[int]bool)
	for idx := 0; idx < len(tIfIdxList); idx++ {
		newTIfIdxMap[int(tIfIdxList[idx])] = true
	}

	vlanEnt, _ := server.vlanPropMap[vlanIfIdx]
	for oldUIfIdx, _ := range vlanEnt.UntagIfIdxMap {
		_, exist := newUIfIdxMap[oldUIfIdx]
		if !exist {
			uIfIdxDelList = append(uIfIdxDelList, int32(oldUIfIdx))
		} else {
			uIfIdxCreateList = append(uIfIdxCreateList, int32(oldUIfIdx))
			delete(newUIfIdxMap, oldUIfIdx)
		}
	}
	for newUIfIdx, _ := range newUIfIdxMap {
		uIfIdxCreateList = append(uIfIdxCreateList, int32(newUIfIdx))
	}
	for oldTIfIdx, _ := range vlanEnt.TagIfIdxMap {
		_, exist := newTIfIdxMap[oldTIfIdx]
		if !exist {
			tIfIdxDelList = append(tIfIdxDelList, int32(oldTIfIdx))
		} else {
			tIfIdxCreateList = append(tIfIdxCreateList, int32(oldTIfIdx))
			delete(newTIfIdxMap, oldTIfIdx)
		}
	}
	for newTIfIdx, _ := range newTIfIdxMap {
		tIfIdxCreateList = append(tIfIdxCreateList, int32(newTIfIdx))
	}
	server.processVlanDelete(vlanName, vlanIfIdx, uIfIdxDelList, tIfIdxDelList)
	server.processVlanCreate(vlanName, vlanIfIdx, uIfIdxCreateList, tIfIdxCreateList)
}

func (server *ARPServer) processLagCreate(lagIfIdx int, portList []int32) {

}

func (server *ARPServer) processLagDelete(lagIfIdx int, portList []int32) {

}

func (server *ARPServer) processLagUpdate(lagIfIdx int, portList []int32) {

}
