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
	_ "errors"
	_ "l3/vrrp/api"
	"l3/vrrp/config"
	_ "l3/vrrp/debug"
	"vrrpd"
)

func (h *ConfigHandler) convertVrrpV4IntfEntryToThriftEntry(state config.IntfCfg) *vrrpd.VrrpV4IntfState {
	entry := vrrpd.NewVrrpV4IntfState()
	return entry
}

func (h *ConfigHandler) convertVrrpV6IntfEntryToThriftEntry(state config.IntfCfg) *vrrpd.VrrpV6IntfState {
	entry := vrrpd.NewVrrpV6IntfState()
	return entry
}

func (h *ConfigHandler) GetBulkVrrpV4IntfState(fromIndex vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpV4IntfStateGetInfo, error) {
	return nil, nil
}

func (h *ConfigHandler) GetVrrpV4IntfState(intfRef string, vrId int32) (*vrrpd.VrrpV4IntfState, error) {
	return nil, nil
}

func (h *ConfigHandler) GetBulkVrrpV6IntfState(fromIndex vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpV6IntfStateGetInfo, error) {
	return nil, nil
}

func (h *ConfigHandler) GetVrrpV6IntfState(intfRef string, vrId int32) (*vrrpd.VrrpV6IntfState, error) {
	return nil, nil
}

func (h *ConfigHandler) GetBulkVrrpStatsState(fromIndex vrrpd.Int, count vrrpd.Int) (*vrrpd.VrrpStatsStateGetInfo, error) {
	return nil, nil
}

func (h *ConfigHandler) GetVrrpStatsState(intfRef string, vrid int32) (*vrrpd.VrrpStatsState, error) {
	return nil, nil
}
