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

package api

import (
	//"errors"
	"l3/ospfv2/objects"
	//"ospfv2d/server"
)

/*
var svr *server.OpticdServer

//Initialize server handle
func InitApiLayer(server *server.OSPFV2Server) {
	svr = server
}
*/

func CreateOspfv2Area(cfg *objects.Ospfv2Area) (bool, error) {
	return true, nil
}

func UpdateOspfv2Area(oldCfg, newCfg *objects.Ospfv2Area, attrset []bool) (bool, error) {
	return true, nil
}

func DeleteOspfv2Area(cfg *objects.Ospfv2Area) (bool, error) {
	return true, nil
}

func GetOspfv2AreaState(areaId uint32) (*objects.Ospfv2AreaState, error) {
	return nil, nil
}

func GetBulkOspfv2AreaState(fromIdx, count int) (*objects.Ospfv2AreaStateGetInfo, error) {
	return nil, nil
}

func CreateOspfv2Global(cfg *objects.Ospfv2Global) (bool, error) {
	return true, nil
}

func UpdateOspfv2Global(oldCfg, newCfg *objects.Ospfv2Global, attrset []bool) (bool, error) {
	return true, nil
}

func DeleteOspfv2Global(cfg *objects.Ospfv2Global) (bool, error) {
	return true, nil
}

func GetOspfv2GlobalState(vrf string) (*objects.Ospfv2GlobalState, error) {
	return nil, nil
}

func GetBulkOspfv2GlobalState(fromIdx, count int) (*objects.Ospfv2GlobalStateGetInfo, error) {
	return nil, nil
}

func CreateOspfv2Intf(cfg *objects.Ospfv2Intf) (bool, error) {
	return true, nil
}

func UpdateOspfv2Intf(oldCfg, newCfg *objects.Ospfv2Intf, attrset []bool) (bool, error) {
	return true, nil
}

func DeleteOspfv2Intf(cfg *objects.Ospfv2Intf) (bool, error) {
	return true, nil
}

func GetOspfv2IntfState(ipAddr, addrLessIfIdx uint32) (*objects.Ospfv2IntfState, error) {
	return nil, nil
}

func GetBulkOspfv2IntfState(fromIdx, count int) (*objects.Ospfv2IntfStateGetInfo, error) {
	return nil, nil
}

func GetOspfv2LsdbState(lsType uint8, lsId, areaId, advRtrId uint32) (*objects.Ospfv2LsdbState, error) {
	return nil, nil
}

func GetBulkOspfv2LsdbState(fromIdx, count int) (*objects.Ospfv2LsdbStateGetInfo, error) {
	return nil, nil
}

func GetOspfv2NbrState(ipAddr, addrLessIfIdx uint32) (*objects.Ospfv2NbrState, error) {
	return nil, nil
}

func GetBulkOspfv2NbrState(fromIdx, count int) (*objects.Ospfv2NbrStateGetInfo, error) {
	return nil, nil
}

func GetOspfv2RouteState(destId, addrMask, destType uint32) (*objects.Ospfv2RouteState, error) {
	return nil, nil
}

func GetBulkOspfv2RouteState(fromIdx, count int) (*objects.Ospfv2RouteStateGetInfo, error) {
	return nil, nil
}
