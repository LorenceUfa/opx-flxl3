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

func (server *OSPFV2Server) initLsdbData() {
	server.LsdbData.AreaLsdb = make(map[LsdbKey]LSDatabase)
	server.LsdbData.AreaSelfOrigLsa = make(map[LsdbKey]SelfOrigLsa)
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
