package fsArpdClnt

import (
	"arpd"
	"models/objects"
)

func convertToThriftFromArpEntryState(obj *objects.ArpEntryState) *arpd.ArpEntryState {
	return &arpd.ArpEntryState{
		Vlan:           string(obj.Vlan),
		MacAddr:        string(obj.MacAddr),
		Intf:           string(obj.Intf),
		IpAddr:         string(obj.IpAddr),
		ExpiryTimeLeft: string(obj.ExpiryTimeLeft),
	}
}

func convertFromThriftToArpEntryState(obj *arpd.ArpEntryState) *objects.ArpEntryState {
	return &objects.ArpEntryState{
		Vlan:           string(obj.Vlan),
		MacAddr:        string(obj.MacAddr),
		Intf:           string(obj.Intf),
		IpAddr:         string(obj.IpAddr),
		ExpiryTimeLeft: string(obj.ExpiryTimeLeft),
	}
}

func convertToThriftToArpLinuxEntryState(obj *objects.ArpLinuxEntryState) *arpd.ArpLinuxEntryState {
	return &arpd.ArpLinuxEntryState{
		HWType:  string(obj.HWType),
		IfName:  string(obj.IfName),
		MacAddr: string(obj.MacAddr),
		IpAddr:  string(obj.IpAddr),
	}
}

func convertFromThriftToArpLinuxEntryState(obj *arpd.ArpLinuxEntryState) *objects.ArpLinuxEntryState {
	return &objects.ArpLinuxEntryState{
		HWType:  string(obj.HWType),
		IfName:  string(obj.IfName),
		MacAddr: string(obj.MacAddr),
		IpAddr:  string(obj.IpAddr),
	}
}

func convertToThriftFromArpGlobal(obj *objects.ArpGlobal) *arpd.ArpGlobal {
	return &arpd.ArpGlobal{
		Vrf:     string(obj.Vrf),
		Timeout: int32(obj.Timeout),
	}
}

func convertFromThriftToArpGlobal(obj *arpd.ArpGlobal) *objects.ArpGlobal {
	return &objects.ArpGlobal{
		Vrf:     string(obj.Vrf),
		Timeout: int32(obj.Timeout),
	}
}
