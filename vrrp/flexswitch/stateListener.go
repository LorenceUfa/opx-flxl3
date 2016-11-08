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

package flexswitch

import (
	"errors"
	"fmt"
	"l3/vrrp/api"
	"l3/vrrp/config"
	"l3/vrrp/debug"
	"vrrpd"
)

func (h *ConfigHandler) convertVrrpV4IntfEntryToThriftEntry(state config.State) *vrrpd.VrrpV4IntfState {
	entry := vrrpd.NewVrrpV4IntfState()
	entry.IntfRef = state.IntfRef
	entry.VRID = state.Vrid
	entry.CurrentState = state.CurrentFsmState
	entry.MasterIp = state.MasterIp
	entry.AdverRx = int32(state.AdverRx)
	entry.AdverTx = int32(state.AdverTx)
	entry.LastAdverRx = state.LastAdverRx
	entry.LastAdverTx = state.LastAdverTx
	entry.IntfIpAddr = state.IpAddr
	entry.Address = state.VirtualIp
	entry.VirtualRouterMACAddress = state.VirtualRouterMACAddress
	entry.MasterDownTimer = state.MasterDownTimer
	return entry
}

func (h *ConfigHandler) convertVrrpV6IntfEntryToThriftEntry(state config.State) *vrrpd.VrrpV6IntfState {
	entry := vrrpd.NewVrrpV6IntfState()
	entry.IntfRef = state.IntfRef
	entry.VRID = state.Vrid
	entry.CurrentState = state.CurrentFsmState
	entry.MasterIp = state.MasterIp
	entry.AdverRx = int32(state.AdverRx)
	entry.AdverTx = int32(state.AdverTx)
	entry.LastAdverRx = state.LastAdverRx
	entry.LastAdverTx = state.LastAdverTx
	entry.IntfIpAddr = state.IpAddr
	entry.Address = state.VirtualIp
	entry.VirtualRouterMACAddress = state.VirtualRouterMACAddress
	entry.MasterDownTimer = state.MasterDownTimer
	return entry
}

func (h *ConfigHandler) GetBulkVrrpV4IntfState(fromIdx vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpV4IntfStateGetInfo, error) {
	debug.Logger.Debug("Get bulk request for vrrp v4 intf states")
	nextIdx, currCount, vrrpEntries := api.GetAllV4IntfStates(int(fromIdx), int(count))
	if len(vrrpEntries) == 0 || vrrpEntries == nil {
		return nil, errors.New("No Vrrp V4 entries configured")
	}
	vrrpResp := make([]*vrrpd.VrrpV4IntfState, len(vrrpEntries))
	for idx, vrrpEntry := range vrrpEntries {
		vrrpResp[idx] = h.convertVrrpV4IntfEntryToThriftEntry(vrrpEntry)
	}
	vrrpEntryBulk := vrrpd.NewVrrpV4IntfStateGetInfo()
	vrrpEntryBulk.StartIdx = fromIdx
	vrrpEntryBulk.EndIdx = vrrpd.Int(nextIdx)
	vrrpEntryBulk.Count = vrrpd.Int(currCount)
	vrrpEntryBulk.More = (nextIdx != 0)
	vrrpEntryBulk.VrrpV4IntfStateList = vrrpResp
	return vrrpEntryBulk, nil
}

func (h *ConfigHandler) GetVrrpV4IntfState(intfRef string, vrId int32) (*vrrpd.VrrpV4IntfState, error) {
	entry := api.GetVrrpIntfEntry(intfRef, vrId, config.VERSION2)
	if entry == nil {
		return nil, errors.New(fmt.Sprintln("No vrrp interface configurea for intfRef:", intfRef, "vrid:", vrId))
	}
	return h.convertVrrpV4IntfEntryToThriftEntry(*entry), nil
}

func (h *ConfigHandler) GetBulkVrrpV6IntfState(fromIdx vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpV6IntfStateGetInfo, error) {
	debug.Logger.Debug("Get bulk request for vrrp v6 intf states")
	nextIdx, currCount, vrrpEntries := api.GetAllV6IntfStates(int(fromIdx), int(count))
	if len(vrrpEntries) == 0 || vrrpEntries == nil {
		return nil, errors.New("No Vrrp V6 entries configured")
	}
	vrrpResp := make([]*vrrpd.VrrpV6IntfState, len(vrrpEntries))
	for idx, vrrpEntry := range vrrpEntries {
		vrrpResp[idx] = h.convertVrrpV6IntfEntryToThriftEntry(vrrpEntry)
	}
	vrrpEntryBulk := vrrpd.NewVrrpV6IntfStateGetInfo()
	vrrpEntryBulk.StartIdx = fromIdx
	vrrpEntryBulk.EndIdx = vrrpd.Int(nextIdx)
	vrrpEntryBulk.Count = vrrpd.Int(currCount)
	vrrpEntryBulk.More = (nextIdx != 0)
	vrrpEntryBulk.VrrpV6IntfStateList = vrrpResp
	return vrrpEntryBulk, nil
}

func (h *ConfigHandler) GetVrrpV6IntfState(intfRef string, vrId int32) (*vrrpd.VrrpV6IntfState, error) {
	entry := api.GetVrrpIntfEntry(intfRef, vrId, config.VERSION3)
	if entry == nil {
		return nil, errors.New(fmt.Sprintln("No vrrp interface configurea for intfRef:", intfRef, "vrid:", vrId))
	}
	return h.convertVrrpV6IntfEntryToThriftEntry(*entry), nil
}
