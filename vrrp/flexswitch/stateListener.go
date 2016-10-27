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
	"l3/vrrp/api"
	"l3/vrrp/config"
	"l3/vrrp/debug"
	"vrrpd"
)

func (h *ConfigHandler) convertVrrpV4IntfEntryToThriftEntry(state vrrpd.VrrpV4IntfState) *vrrpd.VrrpV4IntfState {
	entry := vrrpd.NewVrrpV4IntfState()
	entry.VirtualRouterMACAddress = state.VirtualRouterMACAddress
	entry.PreemptMode = bool(state.PreemptMode)
	entry.AdvertisementInterval = int32(state.AdvertisementInterval)
	entry.VRID = int32(state.VRID)
	entry.Priority = int32(state.Priority)
	entry.SkewTime = int32(state.SkewTime)
	entry.VirtualIPv4Addr = state.VirtualIPv4Addr
	//entry.IfIndex = int32(state.IfIndex)
	entry.MasterDownTimer = int32(state.MasterDownTimer)
	entry.IntfIpAddr = state.IntfIpAddr
	entry.VrrpState = state.VrrpState
	return entry
}

func (h *ConfigHandler) convertVrrpV6IntfEntryToThriftEntry(state vrrpd.VrrpV6IntfState) *vrrpd.VrrpV6IntfState {
	entry := vrrpd.NewVrrpV4IntfState()
	entry.VirtualRouterMACAddress = state.VirtualRouterMACAddress
	entry.PreemptMode = bool(state.PreemptMode)
	entry.AdvertisementInterval = int32(state.AdvertisementInterval)
	entry.VRID = int32(state.VRID)
	entry.Priority = int32(state.Priority)
	entry.SkewTime = int32(state.SkewTime)
	entry.VirtualIPv4Addr = state.VirtualIPv4Addr
	//entry.IfIndex = int32(state.IfIndex)
	entry.MasterDownTimer = int32(state.MasterDownTimer)
	entry.IntfIpAddr = state.IntfIpAddr
	entry.VrrpState = state.VrrpState
	return entry
}

func (h *ConfigHandler) GetBulkVrrpV4IntfState(fromIndex vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpV4IntfStateGetInfo, error) {
	/*
		nextIdx, currCount, vrrpIntfStateEntries := h.server.VrrpGetBulkVrrpIntfStates(int(fromIndex), int(count))
		if vrrpIntfStateEntries == nil {
			return nil, errors.New("Interface Slice is not initialized")
		}
		vrrpEntryResponse := make([]*vrrpd.VrrpIntfState, len(vrrpIntfStateEntries))
		for idx, item := range vrrpIntfStateEntries {
			vrrpEntryResponse[idx] = h.convertVrrpV4IntfEntryToThriftEntry(item)
		}
		intfEntryBulk := vrrpd.NewVrrpV4IntfStateGetInfo()
		intfEntryBulk.VrrpIntfStateList = vrrpEntryResponse
		intfEntryBulk.StartIdx = fromIndex
		intfEntryBulk.EndIdx = vrrpd.Int(nextIdx)
		intfEntryBulk.Count = vrrpd.Int(currCount)
		intfEntryBulk.More = (nextIdx != 0)
		return intfEntryBulk, nil
	*/
	return nil, nil
}

func (h *ConfigHandler) GetVrrpV4IntfState(ifIndex int32, vrId int32) (*vrrpd.VrrpV4IntfState, error) {
	/*
		response := vrrpd.NewVrrpV4IntfState()
		key := strconv.Itoa(int(ifIndex)) + "_" + strconv.Itoa(int(vrId))
		rv := h.server.VrrpPopulateIntfState(key, response)
		if !rv {
			return nil, errors.New(VRRP_SVR_NO_ENTRY + strconv.Itoa(int(ifIndex)) +
				" and Virtual Router Id:" + strconv.Itoa(int(vrId)))
		}
		return response, nil
	*/
	return nil, nil
}

func (h *ConfigHandler) GetBulkVrrpV6IntfState(fromIndex vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpV6IntfStateGetInfo, error) {
	/*
		nextIdx, currCount, vrrpIntfStateEntries := h.server.VrrpGetBulkVrrpIntfStates(int(fromIndex), int(count))
		if vrrpIntfStateEntries == nil {
			return nil, errors.New("Interface Slice is not initialized")
		}
		vrrpEntryResponse := make([]*vrrpd.VrrpIntfState, len(vrrpIntfStateEntries))
		for idx, item := range vrrpIntfStateEntries {
			vrrpEntryResponse[idx] = h.convertVrrpV4IntfEntryToThriftEntry(item)
		}
		intfEntryBulk := vrrpd.NewVrrpV4IntfStateGetInfo()
		intfEntryBulk.VrrpIntfStateList = vrrpEntryResponse
		intfEntryBulk.StartIdx = fromIndex
		intfEntryBulk.EndIdx = vrrpd.Int(nextIdx)
		intfEntryBulk.Count = vrrpd.Int(currCount)
		intfEntryBulk.More = (nextIdx != 0)
		return intfEntryBulk, nil
	*/
	return nil, nil
}

func (h *ConfigHandler) GetVrrpV6IntfState(ifIndex int32, vrId int32) (*vrrpd.VrrpV6IntfState, error) {
	/*
		response := vrrpd.NewVrrpV4IntfState()
		key := strconv.Itoa(int(ifIndex)) + "_" + strconv.Itoa(int(vrId))
		rv := h.server.VrrpPopulateIntfState(key, response)
		if !rv {
			return nil, errors.New(VRRP_SVR_NO_ENTRY + strconv.Itoa(int(ifIndex)) +
				" and Virtual Router Id:" + strconv.Itoa(int(vrId)))
		}
		return response, nil
	*/
	return nil, nil
}
