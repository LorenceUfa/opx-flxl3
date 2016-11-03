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
	"l3/ospfv2/objects"
)

type GlobalStruct struct {
	Vrf                string
	RouterId           uint32
	AdminState         bool
	ASBdrRtrStatus     bool
	ReferenceBandwidth uint32
	AreaBdrRtrStatus   bool
	//isABR             bool
}

func (server *OSPFV2Server) updateGlobal(newCfg, oldCfg *objects.Ospfv2Global, attrset []bool) (bool, error) {
	server.logger.Info("Global configuration update")
	return true, nil
}

func (server *OSPFV2Server) createGlobal(cfg *objects.Ospfv2Global) (bool, error) {
	server.logger.Info("Global configuration create")
	return true, nil
}

func (server *OSPFV2Server) deleteGlobal(cfg *objects.Ospfv2Global) (bool, error) {
	server.logger.Info("Global configuration delete")
	return true, nil
}

func (server *OSPFV2Server) getGlobalState(vrf string) (*objects.Ospfv2GlobalState, error) {
	var retObj objects.Ospfv2GlobalState
	return &retObj, nil
}

func (server *OSPFV2Server) getBulkGlobalState(fromIdx, cnt int) (*objects.Ospfv2GlobalStateGetInfo, error) {
	var retObj objects.Ospfv2GlobalStateGetInfo
	return &retObj, nil
}

/*
func (server *OSPFV2Server) initOspfGlobal() {
	server.globalData.Vrf = "Default"
	server.globalData.RouterId = 0
	server.globalData.AdminState = false
	server.globalData.ASBdrRtrStatus = false
	server.globalData.ReferenceBandwidth = 100000 //Default value 100Gbps
	server.globalData.AreaBdrRtrStatus = false
	server.logger.Info("Global configuration initialized")
}

func (server *OSPFServer) processASBdrRtrStatus(isASBR bool) {
	if isASBR {
		server.logger.Info(fmt.Sprintln("GLOBAL: Router is ASBR. Listen to RIBD updates."))
		//get ribd routes
		//server.testASExternal()
		server.getRibdRoutes()
		server.startRibdUpdates()
	}
}
func (server *OSPFServer) processGlobalConfig(gConf config.GlobalConf) error {
	var localIntfStateMap = make(map[IntfConfKey]config.Status)
	for key, ent := range server.IntfConfMap {
		localIntfStateMap[key] = ent.IfAdminStat
		if ent.IfAdminStat == config.Enabled &&
			server.ospfGlobalConf.AdminStat == config.Enabled {
			server.StopSendRecvPkts(key)
		}
	}

	if server.ospfGlobalConf.AdminStat == config.Enabled {
		server.nbrFSMCtrlCh <- false
		//	server.neighborConfStopCh <- true
		//server.NeighborListMap = nil
		server.StopLSDatabase()
		server.ospfRxNbrPktStopCh <- true
		server.ospfTxNbrPktStopCh <- true
		server.neighborFSMCtrlCh <- false
		server.neighborConfStopCh <- true

	}
	server.logger.Info(fmt.Sprintln("Received call for performing Global Configuration", gConf))
	server.updateGlobalConf(gConf)

	if server.ospfGlobalConf.AdminStat == config.Enabled {
		//server.NeighborListMap = make(map[IntfConfKey]list.List)
		server.logger.Info(fmt.Sprintln("Spawn Neighbor state machine"))
		server.InitNeighborStateMachine()
		go server.UpdateNeighborConf()
		go server.ProcessNbrStateMachine()
		go server.ProcessTxNbrPkt()
		go server.ProcessRxNbrPkt()
		server.StartLSDatabase()

	}
	server.processASBdrRtrStatus(server.ospfGlobalConf.AreaBdrRtrStatus)
	for key, ent := range localIntfStateMap {
		if ent == config.Enabled &&
			server.ospfGlobalConf.AdminStat == config.Enabled {
			server.logger.Info(fmt.Sprintln("Start rx/tx thread."))
			server.StartSendRecvPkts(key)
		}
	}

	return nil
}
*/
