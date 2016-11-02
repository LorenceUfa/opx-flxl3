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

package rpc

import (
	"l3/ospfv2/api"
	"ospfv2d"
)

func (rpcHdl *rpcServiceHandler) CreateOspfv2Intf(config *ospfv2d.Ospfv2Intf) (bool, error) {
	cfg := convertFromRPCFmtOspfv2Intf(config)
	rv, err := api.CreateOspfv2Intf(cfg)
	return rv, err
}

func (rpcHdl *rpcServiceHandler) UpdateOspfv2Intf(oldConfig, newConfig *ospfv2d.Ospfv2Intf, attrset []bool, op []*ospfv2d.PatchOpInfo) (bool, error) {
	convOldCfg := convertFromRPCFmtOspfv2Intf(oldConfig)
	convNewCfg := convertFromRPCFmtOspfv2Intf(newConfig)
	rv, err := api.UpdateOspfv2Intf(convOldCfg, convNewCfg, attrset)
	return rv, err
}

func (rpcHdl *rpcServiceHandler) DeleteOspfv2Intf(config *ospfv2d.Ospfv2Intf) (bool, error) {
	cfg := convertFromRPCFmtOspfv2Intf(config)
	rv, err := api.DeleteOspfv2Intf(cfg)
	return rv, err
}

func (rpcHdl *rpcServiceHandler) GetOspfv2IntfState(IpAddress string, AddressLessIfIdx int32) (*ospfv2d.Ospfv2IntfState, error) {
	var convObj *ospfv2d.Ospfv2IntfState
	//TODO
	ipAddr := uint32(0)
	addrLessIfIdx := uint32(0)
	obj, err := api.GetOspfv2IntfState(ipAddr, addrLessIfIdx)
	if err == nil {
		convObj = convertToRPCFmtOspfv2IntfState(obj)
	}
	return convObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkOspfv2IntfState(fromIdx, count ospfv2d.Int) (*ospfv2d.Ospfv2IntfStateGetInfo, error) {
	var getBulkInfo ospfv2d.Ospfv2IntfStateGetInfo
	info, err := api.GetBulkOspfv2IntfState(int(fromIdx), int(count))
	getBulkInfo.StartIdx = fromIdx
	getBulkInfo.EndIdx = ospfv2d.Int(info.EndIdx)
	getBulkInfo.More = info.More
	getBulkInfo.Count = ospfv2d.Int(len(info.List))
	for idx := 0; idx < len(info.List); idx++ {
		getBulkInfo.Ospfv2IntfStateList = append(getBulkInfo.Ospfv2IntfStateList,
			convertToRPCFmtOspfv2IntfState(info.List[idx]))
	}
	return &getBulkInfo, err
}
