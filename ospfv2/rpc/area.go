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

func (rpcHdl *rpcServiceHandler) CreateOspfv2Area(config *ospfv2d.Ospfv2Area) (bool, error) {
	cfg := convertFromRPCFmtOspfv2Area(config)
	rv, err := api.CreateOspfv2Area(cfg)
	return rv, err

}

func (rpcHdl *rpcServiceHandler) UpdateOspfv2Area(oldConfig, newConfig *ospfv2d.Ospfv2Area, attrset []bool, op []*ospfv2d.PatchOpInfo) (bool, error) {
	convOldCfg := convertFromRPCFmtOspfv2Area(oldConfig)
	convNewCfg := convertFromRPCFmtOspfv2Area(newConfig)
	rv, err := api.UpdateOspfv2Area(convOldCfg, convNewCfg, attrset)
	return rv, err
}

func (rpcHdl *rpcServiceHandler) DeleteOspfv2Area(config *ospfv2d.Ospfv2Area) (bool, error) {
	cfg := convertFromRPCFmtOspfv2Area(config)
	rv, err := api.DeleteOspfv2Area(cfg)
	return rv, err
}

func (rpcHdl *rpcServiceHandler) GetOspfv2AreaState(AreaId string) (*ospfv2d.Ospfv2AreaState, error) {
	var convObj *ospfv2d.Ospfv2AreaState
	//TODO
	areaId := uint32(0)
	obj, err := api.GetOspfv2AreaState(areaId)
	if err == nil {
		convObj = convertToRPCFmtOspfv2AreaState(obj)
	}
	return convObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkOspfv2AreaState(fromIdx, count ospfv2d.Int) (*ospfv2d.Ospfv2AreaStateGetInfo, error) {
	var getBulkInfo ospfv2d.Ospfv2AreaStateGetInfo
	info, err := api.GetBulkOspfv2AreaState(int(fromIdx), int(count))
	getBulkInfo.StartIdx = fromIdx
	getBulkInfo.EndIdx = ospfv2d.Int(info.EndIdx)
	getBulkInfo.More = info.More
	getBulkInfo.Count = ospfv2d.Int(len(info.List))
	for idx := 0; idx < len(info.List); idx++ {
		getBulkInfo.Ospfv2AreaStateList = append(getBulkInfo.Ospfv2AreaStateList,
			convertToRPCFmtOspfv2AreaState(info.List[idx]))
	}
	return &getBulkInfo, err
}
