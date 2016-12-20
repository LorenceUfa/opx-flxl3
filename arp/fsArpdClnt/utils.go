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

package fsArpdClnt

import (
	"arpd"
	//"arpdInt"
	"utils/clntUtils/clntDefs/arpdClntDefs"
)

func convertFromThriftArpEntryState(obj *arpd.ArpEntryState) *arpdClntDefs.ArpEntryState {
	return &arpdClntDefs.ArpEntryState{
		IpAddr:         obj.IpAddr,
		MacAddr:        obj.MacAddr,
		Vlan:           obj.Vlan,
		Intf:           obj.Intf,
		ExpiryTimeLeft: obj.ExpiryTimeLeft,
	}
}

func convertFromThriftArpLinuxEntryState(obj *arpd.ArpLinuxEntryState) *arpdClntDefs.ArpLinuxEntryState {
	return &arpdClntDefs.ArpLinuxEntryState{
		IpAddr:  obj.IpAddr,
		HWType:  obj.HWType,
		MacAddr: obj.MacAddr,
		IfName:  obj.IfName,
	}
}

func convertToThriftArpDeleteByIfName(cfg *arpdClntDefs.ArpDeleteByIfName) *arpd.ArpDeleteByIfName {
	return &arpd.ArpDeleteByIfName{
		IfName: cfg.IfName,
	}
}

func convertToThriftArpDeleteByIPv4Addr(cfg *arpdClntDefs.ArpDeleteByIPv4Addr) *arpd.ArpDeleteByIPv4Addr {
	return &arpd.ArpDeleteByIPv4Addr{
		IpAddr: cfg.IpAddr,
	}
}

func convertToThriftArpRefreshByIfName(cfg *arpdClntDefs.ArpRefreshByIfName) *arpd.ArpRefreshByIfName {
	return &arpd.ArpRefreshByIfName{
		IfName: cfg.IfName,
	}
}

func convertToThriftArpRefreshByIPv4Addr(cfg *arpdClntDefs.ArpRefreshByIPv4Addr) *arpd.ArpRefreshByIPv4Addr {
	return &arpd.ArpRefreshByIPv4Addr{
		IpAddr: cfg.IpAddr,
	}
}

func convertToThriftArpGlobal(cfg *arpdClntDefs.ArpGlobal) *arpd.ArpGlobal {
	return &arpd.ArpGlobal{
		Vrf:     cfg.Vrf,
		Timeout: cfg.Timeout,
	}
}

func convertToThriftPatchOpInfo(oper []*arpdClntDefs.PatchOpInfo) []*arpd.PatchOpInfo {
	var retObj []*arpd.PatchOpInfo
	for _, op := range oper {
		convOp := &arpd.PatchOpInfo{
			Op:    op.Op,
			Path:  op.Path,
			Value: op.Value,
		}
		retObj = append(retObj, convOp)
	}
	return retObj
}
