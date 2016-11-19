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

package server

import (
	"fmt"
	"models/objects"
	"ospfv2d"
)

func (server *OSPFV2Server) StartDBClient() {
	for {
		select {
		case msg := <-server.MessagingChData.RouteTblToDBClntChData.RouteAddMsgCh:
			server.RouteAddToDB(msg)
		case msg := <-server.MessagingChData.RouteTblToDBClntChData.RouteDelMsgCh:
			server.RouteDelFromDB(msg)
		}
	}
}

func (server *OSPFV2Server) RouteAddToDB(msg RouteAddMsg) {
	var dbObj objects.Ospfv2RouteState
	obj := ospfv2d.NewOspfv2RouteState()

	obj.DestId = convertUint32ToDotNotation(msg.RTblKey.DestId)
	obj.AddrMask = convertUint32ToDotNotation(msg.RTblKey.AddrMask)
	obj.DestType = string(msg.RTblKey.DestType)
	obj.OptCapabilities = int32(msg.RTblEntry.RoutingTblEnt.OptCapabilities)
	obj.AreaId = convertUint32ToDotNotation(msg.RTblEntry.AreaId)
	obj.PathType = string(msg.RTblEntry.RoutingTblEnt.PathType)
	obj.Cost = int32(msg.RTblEntry.RoutingTblEnt.Cost)
	obj.Type2Cost = int32(msg.RTblEntry.RoutingTblEnt.Type2Cost)
	obj.NumOfPaths = int16(msg.RTblEntry.RoutingTblEnt.NumOfPaths)
	nh_list := make([]ospfv2d.Ospfv2NextHop, len(msg.RTblEntry.RoutingTblEnt.NextHops))
	idx := 0
	for nxtHop, _ := range msg.RTblEntry.RoutingTblEnt.NextHops {
		nh_list[idx].IntfIPAddr = convertUint32ToDotNotation(nxtHop.IfIPAddr)
		nh_list[idx].IntfIdx = int32(nxtHop.IfIdx)
		nh_list[idx].NextHopIPAddr = convertUint32ToDotNotation(nxtHop.NextHopIP)
		nh_list[idx].AdvRtrId = convertUint32ToDotNotation(nxtHop.AdvRtr)
		obj.NextHops = append(obj.NextHops, &nh_list[idx])
		idx++
	}
	obj.LSOrigin = &ospfv2d.Ospfv2LsaKey{}
	obj.LSOrigin.LSType = int8(msg.RTblEntry.RoutingTblEnt.LSOrigin.LSType)
	obj.LSOrigin.LSId = convertUint32ToDotNotation(msg.RTblEntry.RoutingTblEnt.LSOrigin.LSId)
	obj.LSOrigin.AdvRouter = convertUint32ToDotNotation(msg.RTblEntry.RoutingTblEnt.LSOrigin.AdvRouter)
	objects.ConvertThriftToospfv2dOspfv2RouteStateObj(obj, &dbObj)
	if server.dbHdl == nil {
		server.logger.Err("Db Handler is nil")
		return
	}
	err := server.dbHdl.StoreObjectInDb(dbObj)
	if err != nil {
		server.logger.Err(fmt.Sprintln("Failed to add route in db:", err))
	}
	return
}

func (server *OSPFV2Server) RouteDelFromDB(msg RouteDelMsg) {
	var dbObj objects.Ospfv2RouteState
	obj := ospfv2d.NewOspfv2RouteState()

	obj.LSOrigin = &ospfv2d.Ospfv2LsaKey{}
	obj.DestId = convertUint32ToDotNotation(msg.RTblKey.DestId)
	obj.AddrMask = convertUint32ToDotNotation(msg.RTblKey.AddrMask)
	obj.DestType = string(msg.RTblKey.DestType)

	objects.ConvertThriftToospfv2dOspfv2RouteStateObj(obj, &dbObj)
	if server.dbHdl == nil {
		server.logger.Err("Db Handler is nil")
		return
	}
	err := server.dbHdl.DeleteObjectFromDb(dbObj)
	if err != nil {
		server.logger.Err(fmt.Sprintln("Failed to delete route in db:", err))
	}
	return
}
