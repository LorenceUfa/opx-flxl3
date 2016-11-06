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
	"errors"
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

func genOspfv2GlobalUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0

	if attrset == nil {
		mask = objects.OSPFV2_GLOBAL_UPDATE_ROUTER_ID |
			objects.OSPFV2_GLOBAL_UPDATE_ADMIN_STATE |
			objects.OSPFV2_GLOBAL_UPDATE_AS_BDR_RTR_STATUS |
			objects.OSPFV2_GLOBAL_UPDATE_REFERENCE_BANDWIDTH
	} else {
		for idx, val := range attrset {
			if true == val {
				switch idx {
				case 0:
					// Vrf
				case 1:
					mask |= objects.OSPFV2_GLOBAL_UPDATE_ROUTER_ID
				case 2:
					mask |= objects.OSPFV2_GLOBAL_UPDATE_ADMIN_STATE
				case 3:
					mask |= objects.OSPFV2_GLOBAL_UPDATE_AS_BDR_RTR_STATUS
				case 4:
					mask |= objects.OSPFV2_GLOBAL_UPDATE_REFERENCE_BANDWIDTH
				}
			}
		}
	}
	return mask
}

func (server *OSPFV2Server) updateGlobal(newCfg, oldCfg *objects.Ospfv2Global, attrset []bool) (bool, error) {
	server.logger.Info("Global configuration update")
	if server.globalData.AdminState == true {
		// TODO
		//Stop OSPF Interface FSM
		//Flush LSDB
		//Delete all the routes
		//Flush all the routes
		//Stop Neighbor FSM
		//Stop Ribd updates
	}

	mask := genOspfv2GlobalUpdateMask(attrset)
	if mask&objects.OSPFV2_GLOBAL_UPDATE_ADMIN_STATE == objects.OSPFV2_GLOBAL_UPDATE_ADMIN_STATE {
		server.globalData.AdminState = newCfg.AdminState
	}
	if mask&objects.OSPFV2_GLOBAL_UPDATE_ROUTER_ID == objects.OSPFV2_GLOBAL_UPDATE_ROUTER_ID {
		server.globalData.RouterId = newCfg.RouterId
	}
	if mask&objects.OSPFV2_GLOBAL_UPDATE_AS_BDR_RTR_STATUS == objects.OSPFV2_GLOBAL_UPDATE_AS_BDR_RTR_STATUS {
		server.globalData.ASBdrRtrStatus = newCfg.ASBdrRtrStatus
	}
	if mask&objects.OSPFV2_GLOBAL_UPDATE_REFERENCE_BANDWIDTH == objects.OSPFV2_GLOBAL_UPDATE_REFERENCE_BANDWIDTH {
		server.globalData.ReferenceBandwidth = newCfg.ReferenceBandwidth
	}

	if server.globalData.AdminState == true {
		for intfConfKey, intfConfEnt := range server.IntfConfMap {
			if intfConfEnt.AdminState == true {
				server.logger.Info("Server Interface Key", intfConfKey)
				// TODO
				//Start OSPF Interface FSM
				//Start SPF
				//Start Neighbor FSM
				//Start Ribd Updates if ASBdrRtrStatus = true
			}
		}
	}

	return true, nil
}

func (server *OSPFV2Server) createGlobal(cfg *objects.Ospfv2Global) (bool, error) {
	server.logger.Info("Global configuration create")
	if cfg.Vrf != "Default" {
		server.logger.Err("Vrp other than Default is not supported")
		return false, errors.New("Vrp other than Default is not supported")
	}
	server.globalData.Vrf = cfg.Vrf
	server.globalData.AdminState = cfg.AdminState
	server.globalData.RouterId = cfg.RouterId
	server.globalData.ASBdrRtrStatus = cfg.ASBdrRtrStatus
	server.globalData.ReferenceBandwidth = cfg.ReferenceBandwidth
	if cfg.AdminState == true {
		//Restart Neighbor FSM
		//Flush all the routes
		//Flush LSDB
		//Start OSPF Interface FSM
	} else {
		//Stop OSPF Interface FSM
		//Flush LSDB
		//Flush all the routes
		//Restart Neighbor FSM
	}
	return true, nil
}

func (server *OSPFV2Server) deleteGlobal(cfg *objects.Ospfv2Global) (bool, error) {
	server.logger.Info("Global configuration delete")
	server.logger.Err("Global Configuration delete not supported")
	return false, errors.New("Global Configuration delete not supported")
}

func (server *OSPFV2Server) getGlobalState(vrf string) (*objects.Ospfv2GlobalState, error) {
	var retObj objects.Ospfv2GlobalState
	retObj.Vrf = vrf
	retObj.AreaBdrRtrStatus = server.globalData.AreaBdrRtrStatus
	return &retObj, nil
}

func (server *OSPFV2Server) getBulkGlobalState(fromIdx, cnt int) (*objects.Ospfv2GlobalStateGetInfo, error) {
	var retObj objects.Ospfv2GlobalStateGetInfo
	if fromIdx > 0 {
		return nil, errors.New("Invalid range.")
	}
	retObj.EndIdx = 1
	retObj.More = false
	retObj.Count = 1
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		obj, err := server.getGlobalState("Default")
		if err != nil {
			server.logger.Err("Error getting the Ospfv2GlobalState for vrf=default")
			return nil, err
		}
		retObj.List = append(retObj.List, obj)
	}
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
