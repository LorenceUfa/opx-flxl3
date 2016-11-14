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

func (server *OSPFV2Server) initLsdbData() {
	server.LsdbData.AreaLsdb = make(map[LsdbKey]LSDatabase)
	server.LsdbData.AreaSelfOrigLsa = make(map[LsdbKey]SelfOrigLsa)
	server.LsdbData.LsdbCtrlChData.LsdbCtrlCh = make(chan bool)
	server.LsdbData.LsdbCtrlChData.LsdbCtrlReplyCh = make(chan bool)
	server.LsdbData.AgedLsaData.AgedLsaMap = make(map[AgedLsaKey]bool)
	server.LsdbData.LsdbAgingTicker = nil
}

func (server *OSPFV2Server) dinitLsdb() {
	server.LsdbData.LsdbAgingTicker = nil
	server.LsdbData.LsdbCtrlChData.LsdbCtrlCh = nil
	server.LsdbData.LsdbCtrlChData.LsdbCtrlReplyCh = nil
	server.LsdbData.AgedLsaData.AgedLsaMap = nil
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
func (server *OSPFV2Server) DinitAreaLsdb(areaId uint32) {
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

func (server *OSPFV2Server) StartLsdbRoutine() {
	server.initLsdbData()
	go server.ProcessLsdb()
	//Start LsdbAgingTicker
	server.LsdbData.LsdbAgingTicker = time.NewTicker(LsaAgingTimeGranularity)
}

func (server *OSPFV2Server) StopLsdbRoutine() {
	server.LsdbData.LsdbAgingTicker.Stop()
	server.LsdbData.LsdbCtrlChData.LsdbCtrlCh <- true
	cnt := 0
	for {
		select {
		case _ = <-server.LsdbData.LsdbCtrlChData.LsdbCtrlReplyCh:
			server.logger.Info("Successfully Stopped ProcessLsdb routine")
			server.dinitLsdb()
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
	for {
		select {
		case _ = <-server.LsdbData.LsdbCtrlChData.LsdbCtrlCh:
			server.logger.Info("Stopping ProcessLsdb routine")
			server.LsdbData.LsdbCtrlChData.LsdbCtrlReplyCh <- true
			return
		case msg := <-server.MessagingChData.IntfFSMToLsdbChData.GenerateRouterLSACh:
			server.logger.Info("Generate self originated Router LSA", msg)
			lsaKey, err := server.GenerateRouterLSA(msg)
			if err != nil {
				continue
			}
			server.SendMsgToStartSpf()
			spfState := <-server.MessagingChData.SPFToLsdbChData.DoneSPF
			server.logger.Debug("SPF Calculation Return Status", spfState)
			//Flood
			server.SendMsgToFloodSelfOrigLsa(msg.AreaId, lsaKey)
			if server.globalData.AreaBdrRtrStatus == true {
				//TODO:
				//Summary LSA
			}
		case msg := <-server.MessagingChData.NbrFSMToLsdbChData.UpdateSelfNetworkLSACh:
			server.logger.Info("Update self originated Network LSA", msg)
			areaId, lsaKey, err := server.processUpdateSelfNetworkLSA(msg)
			if err != nil {
				continue
			}
			server.SendMsgToStartSpf()
			spfState := <-server.MessagingChData.SPFToLsdbChData.DoneSPF
			server.logger.Debug("SPF Calculation Return Status", spfState)
			//Flood
			// If op == FLUSH then Add lsa to Max AgedLsaStruct
			// else flood using LsdbToFloodForSelfOrigLSAMsg
			if msg.Op != FLUSH {
				server.SendMsgToFloodSelfOrigLsa(areaId, lsaKey)
			}
			if server.globalData.AreaBdrRtrStatus == true {
				//TODO:
				//Summary LSA
			}
		case msg := <-server.MessagingChData.NbrFSMToLsdbChData.RecvdLsaMsgCh:
			server.logger.Info("Update LSA", msg)
			server.processRecvdLSA(msg)
			server.SendMsgToStartSpf()
			spfState := <-server.MessagingChData.SPFToLsdbChData.DoneSPF
			server.logger.Debug("SPF Calculation Return Status", spfState)
			if server.globalData.AreaBdrRtrStatus == true {
				//TODO:
				//Summary LSA
			}
		case msg := <-server.MessagingChData.NbrFSMToLsdbChData.RecvdSelfLsaMsgCh:
			server.logger.Info("Recvd Self LSA", msg)
		case <-server.LsdbData.LsdbAgingTicker.C:
			server.processLsdbAgingTicker()
		}
	}
}
