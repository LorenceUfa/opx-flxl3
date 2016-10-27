package vxlan

import (
	"net"
)

const (
	VXLANBaseClientStr = "BaseClient"
	VXLANSnapClientStr = "SnapClient"
	VXLANMockClientStr = "SnapMockTestClient"
)

type PortEvtCb func(ifindex int32)

// interface class is used to store the communication methods
// for the various daemon communications
type VXLANClientIntf interface {
	IsClientIntfType(client VXLANClientIntf, clientStr string) bool
	// used to notify server of updates
	SetServerChannels(s *VxLanConfigChannels)
	ConnectToClients(clientFile string)
	ConstructPortConfigMap()
	// create/delete
	CreateVtep(vtep *VtepDbEntry, vteplistener chan<- MachineEvent)
	DeleteVtep(vtep *VtepDbEntry)
	CreateVxlan(vxlan *VxlanDbEntry)
	DeleteVxlan(vxlan *VxlanDbEntry)
	UpdateVxlan(vni uint32, addvlanlist, delvlanlist, addUntaggedVlan, delUntaggedVlan []uint16)
	// access ports
	//AddHostToVxlan(vni int32, intfreflist, untagintfreflist []string)
	//DelHostFromVxlan(vni int32, intfreflist, untagintfreflist []string)
	// vtep fsm
	GetIntfInfo(name string, intfchan chan<- MachineEvent)
	GetNextHopInfo(ip net.IP, nexthopchan chan<- MachineEvent)
	ResolveNextHopMac(nextHopIp net.IP, nexthopif string, nexthopmacchan chan<- MachineEvent)

	GetLinkState(ifname string) string
	GetAllVlans() []uint16
	RegisterLinkUpDownEvents(ifindex int32, upcb PortEvtCb, downdb PortEvtCb)
}

type BaseClientIntf struct {
}

func (b BaseClientIntf) IsClientIntfType(client VXLANClientIntf, clientStr string) bool {
	switch client.(type) {
	case *BaseClientIntf:
		if clientStr == "BaseClient" {
			return true
		}
	}
	return false
}

func (b BaseClientIntf) SetServerChannels(s *VxLanConfigChannels) {

}
func (b BaseClientIntf) ConnectToClients(clientFile string) {

}
func (b BaseClientIntf) ConstructPortConfigMap() {

}
func (b BaseClientIntf) GetIntfInfo(name string, intfchan chan<- MachineEvent) {

}
func (b BaseClientIntf) CreateVtep(vtep *VtepDbEntry, vteplistener chan<- MachineEvent) {

}
func (b BaseClientIntf) DeleteVtep(vtep *VtepDbEntry) {

}
func (b BaseClientIntf) CreateVxlan(vxlan *VxlanDbEntry) {

}
func (b BaseClientIntf) DeleteVxlan(vxlan *VxlanDbEntry) {

}
func (b BaseClientIntf) UpdateVxlan(
	vni uint32, addvlanlist, delvlanlist, addUntaggedVlan, delUntaggedVlan []uint16) {

}
func (b BaseClientIntf) CreateVxlanAccess() {

}
func (b BaseClientIntf) DeleteVxlanAccess() {

}
func (b BaseClientIntf) GetNextHopInfo(ip net.IP, nexthopchan chan<- MachineEvent) {

}
func (b BaseClientIntf) ResolveNextHopMac(nextHopIp net.IP, nextHopIfName string, nexthopmacchan chan<- MachineEvent) {

}
func (b BaseClientIntf) AddHostToVxlan(vni int32, intfreflist, untagintfreflist []string) {

}
func (b BaseClientIntf) DelHostFromVxlan(vni int32, intfreflist, untagintfreflist []string) {

}

func (b BaseClientIntf) GetAllVlans() []uint16 {
	return []uint16{}
}
func (b BaseClientIntf) RegisterLinkUpDownEvents(ifindex int32, upcb PortEvtCb, downdb PortEvtCb) {

}
func (b BaseClientIntf) GetLinkState(ifname string) string {
	return "UP"
}
