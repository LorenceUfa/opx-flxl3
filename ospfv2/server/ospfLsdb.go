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
	"time"
)

func (server *OSPFV2Server) InitLsdbData() {
	server.LsdbData.AreaLsdb = make(map[LsdbKey]LSDatabase)
	server.LsdbData.AreaSelfOrigLsa = make(map[LsdbKey]SelfOrigLsa)
	server.LsdbData.LsdbAgingTicker = nil
}

func (server *OSPFV2Server) DeinitLsdb() {
	server.LsdbData.LsdbAgingTicker = nil
	for lsdbKey, _ := range server.LsdbData.AreaLsdb {
		delete(server.LsdbData.AreaLsdb, lsdbKey)
	}
	for lsdbKey, _ := range server.LsdbData.AreaSelfOrigLsa {
		delete(server.LsdbData.AreaSelfOrigLsa, lsdbKey)
	}
	server.LsdbData.AreaLsdb = nil
	server.LsdbData.AreaSelfOrigLsa = nil
}

func (server *OSPFV2Server) InitAreaLsdb(areaId uint32) {
	server.logger.Debug("LSDB: Initialise LSDB for area id ", areaId)
	lsdbKey := LsdbKey{
		AreaId: areaId,
	}
	lsDbEnt, exist := server.LsdbData.AreaLsdb[lsdbKey]
	if !exist {
		lsDbEnt.RouterLsaMap = make(map[LsaKey]RouterLsa)
		lsDbEnt.NetworkLsaMap = make(map[LsaKey]NetworkLsa)
		lsDbEnt.Summary3LsaMap = make(map[LsaKey]SummaryLsa)
		lsDbEnt.Summary4LsaMap = make(map[LsaKey]SummaryLsa)
		lsDbEnt.ASExternalLsaMap = make(map[LsaKey]ASExternalLsa)
		server.LsdbData.AreaLsdb[lsdbKey] = lsDbEnt
	}
	selfOrigLsaEnt, exist := server.LsdbData.AreaSelfOrigLsa[lsdbKey]
	if !exist {
		selfOrigLsaEnt = make(map[LsaKey]bool)
		server.LsdbData.AreaSelfOrigLsa[lsdbKey] = selfOrigLsaEnt
	}

}
func (server *OSPFV2Server) DeinitAreaLsdb(areaId uint32) {
	lsdbKey := LsdbKey{
		AreaId: areaId,
	}
	lsDbEnt, exist := server.LsdbData.AreaLsdb[lsdbKey]
	if exist {
		lsDbEnt.RouterLsaMap = nil
		lsDbEnt.NetworkLsaMap = nil
		lsDbEnt.Summary3LsaMap = nil
		lsDbEnt.Summary4LsaMap = nil
		lsDbEnt.ASExternalLsaMap = nil
		delete(server.LsdbData.AreaLsdb, lsdbKey)
	}
	_, exist = server.LsdbData.AreaSelfOrigLsa[lsdbKey]
	if exist {
		delete(server.LsdbData.AreaSelfOrigLsa, lsdbKey)
	}
}

func (server *OSPFV2Server) FlushAreaLsdb(areaId uint32) {
	server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlCh <- areaId
	<-server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlReplyCh
}

func (server *OSPFV2Server) StartLsdbRoutine() {
	server.LsdbData.LsdbCtrlChData.LsdbGblCtrlCh = make(chan bool)
	server.LsdbData.LsdbCtrlChData.LsdbGblCtrlReplyCh = make(chan bool)
	server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlCh = make(chan uint32)
	server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlReplyCh = make(chan uint32)
	go server.ProcessLsdb()
}

func (server *OSPFV2Server) StopLsdbRoutine() {
	server.LsdbData.LsdbCtrlChData.LsdbGblCtrlCh <- true
	cnt := 0
	for {
		select {
		case _ = <-server.LsdbData.LsdbCtrlChData.LsdbGblCtrlReplyCh:
			server.logger.Info("Successfully Stopped ProcessLsdb routine")
			server.LsdbData.LsdbCtrlChData.LsdbGblCtrlCh = nil
			server.LsdbData.LsdbCtrlChData.LsdbGblCtrlReplyCh = nil
			server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlCh = nil
			server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlReplyCh = nil
			return
		default:
			time.Sleep(time.Duration(10) * time.Millisecond)
			cnt = cnt + 1
			if cnt == 100 {
				server.logger.Err("Unable to stop the ProcessLsdb routine")
				return
			}
		}
	}
}

func (server *OSPFV2Server) processRecvdLSA(msg RecvdLsaMsg) error {
	switch msg.LsaKey.LSType {
	case RouterLSA:
		server.processRecvdRouterLSA(msg)
	case NetworkLSA:
		server.processRecvdNetworkLSA(msg)
	case Summary3LSA:
		server.processRecvdSummaryLSA(msg)
	case Summary4LSA:
		server.processRecvdSummaryLSA(msg)
	case ASExternalLSA:
		server.processRecvdASExternalLSA(msg)
	default:
		server.logger.Err("Invalid LsaType:", msg)
	}
	return nil
}

func (server *OSPFV2Server) processRecvdSelfLSA(msg RecvdSelfLsaMsg) error {
	switch msg.LsaKey.LSType {
	case RouterLSA:
		server.processRecvdSelfRouterLSA(msg)
	case NetworkLSA:
		server.processRecvdSelfNetworkLSA(msg)
	case Summary3LSA:
		server.processRecvdSelfSummaryLSA(msg)
	case Summary4LSA:
		server.processRecvdSelfSummaryLSA(msg)
	case ASExternalLSA:
		server.processRecvdSelfASExternalLSA(msg)
	default:
		server.logger.Err("Invalid LsaType:", msg)
	}
	return nil
}

func (server *OSPFV2Server) ProcessLsdb() {
	server.InitLsdbData()
	server.LsdbData.LsdbAgingTicker = time.NewTicker(LsaAgingTimeGranularity)
	for {
		select {
		case _ = <-server.LsdbData.LsdbCtrlChData.LsdbGblCtrlCh:
			server.logger.Info("Stopping ProcessLsdb routine")
			server.LsdbData.LsdbAgingTicker.Stop()
			server.DeinitLsdb()
			server.LsdbData.LsdbCtrlChData.LsdbGblCtrlReplyCh <- true
			return
		case areaId := <-server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlCh:
			server.DeinitAreaLsdb(areaId)
			server.CalcSPFAndRoutingTbl()
			server.LsdbData.LsdbCtrlChData.LsdbAreaCtrlReplyCh <- areaId
		case msg := <-server.MessagingChData.IntfFSMToLsdbChData.GenerateRouterLSACh:
			server.logger.Info("Generate self originated Router LSA", msg)
			err := server.GenerateRouterLSA(msg)
			if err != nil {
				continue
			}
			server.CalcSPFAndRoutingTbl()
		case msg := <-server.MessagingChData.NbrFSMToLsdbChData.UpdateSelfNetworkLSACh:
			server.logger.Info("Update self originated Network LSA", msg)
			err := server.processUpdateSelfNetworkLSA(msg)
			if err != nil {
				continue
			}
			server.CalcSPFAndRoutingTbl()
		case msg := <-server.MessagingChData.NbrFSMToLsdbChData.RecvdLsaMsgCh:
			server.logger.Info("Update LSA", msg)
			server.processRecvdLSA(msg)
			server.CalcSPFAndRoutingTbl()
		case msg := <-server.MessagingChData.NbrFSMToLsdbChData.RecvdSelfLsaMsgCh:
			server.logger.Info("Recvd Self LSA", msg)
			server.processRecvdSelfLSA(msg)
			server.CalcSPFAndRoutingTbl()
		//TODO: Handle AS External
		case <-server.LsdbData.LsdbAgingTicker.C:
			server.processLsdbAgingTicker()
		}
	}
}

func (server *OSPFV2Server) CalcSPFAndRoutingTbl() {
	server.SummaryLsDb = nil
	server.SendMsgToStartSpf()
	spfState := <-server.MessagingChData.SPFToLsdbChData.DoneSPF
	server.logger.Debug("SPF Calculation Return Status", spfState)
	if server.globalData.AreaBdrRtrStatus == true {
		server.logger.Info("Examine transit areas, Summary LSA...")
		server.HandleTransitAreaSummaryLsa()
		server.logger.Info("Generate Summary LSA...")
		server.GenerateSummaryLsa()
		server.logger.Info("========", server.SummaryLsDb, "==========")
		//Summary LSA
		server.installSummaryLsa()
	}
}
