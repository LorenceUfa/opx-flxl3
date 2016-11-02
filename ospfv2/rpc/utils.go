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
	"l3/ospfv2/objects"
	"ospfv2d"
)

func convertFromRPCFmtOspfv2Area(config *ospfv2d.Ospfv2Area) *objects.Ospfv2Area {
	return &objects.Ospfv2Area{}
}

func convertToRPCFmtOspfv2AreaState(obj *objects.Ospfv2AreaState) *ospfv2d.Ospfv2AreaState {
	return &ospfv2d.Ospfv2AreaState{}
}

func convertFromRPCFmtOspfv2Global(config *ospfv2d.Ospfv2Global) *objects.Ospfv2Global {
	return &objects.Ospfv2Global{}
}

func convertToRPCFmtOspfv2GlobalState(obj *objects.Ospfv2GlobalState) *ospfv2d.Ospfv2GlobalState {
	return &ospfv2d.Ospfv2GlobalState{}
}

func convertFromRPCFmtOspfv2Intf(config *ospfv2d.Ospfv2Intf) *objects.Ospfv2Intf {
	return &objects.Ospfv2Intf{}
}

func convertToRPCFmtOspfv2IntfState(obj *objects.Ospfv2IntfState) *ospfv2d.Ospfv2IntfState {
	return &ospfv2d.Ospfv2IntfState{}
}

func convertToRPCFmtOspfv2LsdbState(obj *objects.Ospfv2LsdbState) *ospfv2d.Ospfv2LsdbState {
	return &ospfv2d.Ospfv2LsdbState{}
}

func convertToRPCFmtOspfv2NbrState(obj *objects.Ospfv2NbrState) *ospfv2d.Ospfv2NbrState {
	return &ospfv2d.Ospfv2NbrState{}
}

func convertToRPCFmtOspfv2RouteState(obj *objects.Ospfv2RouteState) *ospfv2d.Ospfv2RouteState {
	return &ospfv2d.Ospfv2RouteState{}
}
