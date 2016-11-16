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
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package server

import (
	"l3/ospfv2/objects"
)

func (server *OSPFV2Server) StopAllRxTxPkt() {
	for intfConfKey, intfConfEnt := range server.IntfConfMap {
		if intfConfEnt.FSMState != objects.INTF_FSM_STATE_DOWN {
			server.StopIntfRxTxPkt(intfConfKey)
		}
	}
}

func (server *OSPFV2Server) StopAreaRxTxPkt(areaId uint32) {
	for intfConfKey, intfConfEnt := range server.IntfConfMap {
		if intfConfEnt.AreaId == areaId {
			if intfConfEnt.FSMState != objects.INTF_FSM_STATE_DOWN {
				server.StopIntfRxTxPkt(intfConfKey)
			}
		}
	}
}

func (server *OSPFV2Server) StopIntfRxTxPkt(intfKey IntfConfKey) {
	ent, _ := server.IntfConfMap[intfKey]
	areaConf, err := server.GetAreaConfForGivenArea(ent.AreaId)
	if err != nil {
		server.logger.Err("Error:", err)
		return
	}

	if server.globalData.AdminState == true &&
		ent.AdminState == true &&
		areaConf.AdminState == true &&
		ent.OperState == true {

		server.StopOspfRecvPkts(intfKey)
		//Nothing to stop for Tx
		server.DeinitRxPkt(intfKey)
		server.DeinitTxPkt(intfKey)
	}
}

func (server *OSPFV2Server) StartAllRxTxPkt() {
	for intfConfKey, _ := range server.IntfConfMap {
		server.StartIntfRxTxPkt(intfConfKey)
	}
}

func (server *OSPFV2Server) StartAreaRxTxPkt(areaId uint32) {
	for intfConfKey, intfConfEnt := range server.IntfConfMap {
		if areaId == intfConfEnt.AreaId {
			server.StartIntfRxTxPkt(intfConfKey)
		}
	}
}

func (server *OSPFV2Server) StartIntfRxTxPkt(intfKey IntfConfKey) {
	ent, _ := server.IntfConfMap[intfKey]
	areaConf, err := server.GetAreaConfForGivenArea(ent.AreaId)
	if err != nil {
		server.logger.Err("Error:", err)
		return
	}

	if server.globalData.AdminState == true &&
		ent.AdminState == true &&
		areaConf.AdminState == true &&
		ent.OperState == true {
		err := server.InitRxPkt(intfKey)
		if err != nil {
			server.logger.Err("Error: InitRxPkt()", err)
			return
		}
		err = server.InitTxPkt(intfKey)
		if err != nil {
			server.logger.Err("Error: InitTxPkt()", err)
			return
		}
		go server.StartOspfRecvPkts(intfKey)
	}
}
