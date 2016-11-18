//
//C_yright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a c_y of the License at
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
	//"encoding/binary"
	"fmt"
	//"l3/ospf/config"
	//"time"
)

func (server *OSPFServer) generateDbSummaryList(nbrConfKey NbrDbdDataKey) {
	nbrConf, exists := server.NeighborConfigMap[nbrConfKey]

	if !exists {
		server.logger.Err(fmt.Sprintln("negotiation: db_list Nbr  doesnt exist. nbr ", nbrConfKey))
		return
	}
	intf, _ := server.IntfConfMap[nbrConf.intfConfKey]

	areaId := convertIPv4ToUint32(intf.AreaId)
	lsdbKey := LsdbKey{
		AreaId: areaId,
	}
	area_lsa, exist := server.LsdbData.AreaLsdb[lsdbKey]
	if !exist {
		server.logger.Err(fmt.Sprintln("negotiation: db_list self originated lsas dont exist. Nbr , lsdb_key ", nbrConfKey, lsdbKey))
		return
	}
	router_lsdb := area_lsa.RouterLsaMap
	network_lsa := area_lsa.NetworkLsaMap
	NbrDBSummaryList[nbrConfKey] = nil
	db_list := []*NbrDbSummaryData{}
	for lsaKey, _ := range router_lsdb {
		// check if lsa instance is marked true
		db_summary := newNbrDbSummaryData()
		drlsa, ret := server.LsdbData.getRouterLsaFromLsdb(areaId, lsaKey)
		if ret == LsdbEntryNotFound {
			continue
		}
		db_summary.lsa_headers = getLsaHeaderFromLsa(drlsa.LsaMd.LSAge, drlsa.LsaMd.Options,
			RouterLSA, lsaKey.LSId, lsaKey.AdvRouter,
			uint32(drlsa.LsaMd.LSSequenceNum), drlsa.LsaMd.LSChecksum,
			drlsa.LsaMd.LSLen)
		db_summary.valid = true
		/* add entry to the db summary list  */
		db_list = append(db_list, db_summary)
		//lsid := convertUint32ToIPv4(lsaKey.LSId)
		//server.logger.Info(fmt.Sprintln("negotiation: db_list append router lsid  ", lsid))
	} // end of for

	for networkKey, _ := range network_lsa {
		// check if lsa instance is marked true
		db_summary := newNbrDbSummaryData()
		//if nbrMdata.isDR {
		dnlsa, ret := server.LsdbData.getNetworkLsaFromLsdb(areaId, networkKey)
		if ret == LsdbEntryNotFound {
			continue
		}
		db_summary.lsa_headers = getLsaHeaderFromLsa(dnlsa.LsaMd.LSAge, dnlsa.LsaMd.Options,
			NetworkLSA, networkKey.LSId, networkKey.AdvRouter,
			uint32(dnlsa.LsaMd.LSSequenceNum), dnlsa.LsaMd.LSChecksum,
			dnlsa.LsaMd.LSLen)
		db_summary.valid = true
		/* add entry to the db summary list  */
		db_list = append(db_list, db_summary)

	} // end of for

	/*   attach summary list */

	summary3_list := server.LsdbData.generateDbsummary3LsaList(areaId)
	if summary3_list != nil {
		db_list = append(db_list, summary3_list...)
	}

	summary4_list := server.generateDbsummary4LsaList(areaId)
	if summary4_list != nil {
		db_list = append(db_list, summary4_list...)
	}

	asExternal_list := server.generateDbasExternalList(areaId)
	if asExternal_list != nil {
		db_list = append(db_list, asExternal_list...)
	}

	for lsa := range db_list {
		rtr_id := convertUint32ToIPv4(db_list[lsa].lsa_headers.adv_router_id)
		server.logger.Info(fmt.Sprintln(lsa, ": ", rtr_id, " lsatype ", db_list[lsa].lsa_headers.ls_type))
	}
	NbrDBSummaryList[nbrConfKey] = db_list
}

/*@fn generateDbasExternalList
This function generates As external list if the router is ASBR
*/
func (server *OSPFServer) generateDbasExternalList(self_areaId uint32) []*NbrDbSummaryData {
	if !server.ospfGlobalConf.AreaBdrRtrStatus {
		return nil // dont add self gen LSA if I am not ASBR
	}
	db_list := []*NbrDbSummaryData{}
	lsdbKey := LsdbKey{
		AreaId: self_areaId,
	}

	area_lsa, exist := server.LsdbData.AreaLsdb[lsdbKey]
	if !exist {
		server.logger.Err(fmt.Sprintln("negotiation: As external LSA doesnt exist"))
		return nil
	}
	as_lsdb := area_lsa.ASExternalLsaMap

	for lsaKey, _ := range as_lsdb {
		// check if lsa instance is marked true
		db_as := newNbrDbSummaryData()
		drlsa, ret := server.LsdbData.getASExternalLsaFromLsdb(self_areaId, lsaKey)
		if ret == LsdbEntryNotFound {
			continue
		}
		db_as.lsa_headers = getLsaHeaderFromLsa(drlsa.LsaMd.LSAge, drlsa.LsaMd.Options,
			ASExternalLSA, lsaKey.LSId, lsaKey.AdvRouter,
			uint32(drlsa.LsaMd.LSSequenceNum), drlsa.LsaMd.LSChecksum,
			drlsa.LsaMd.LSLen)
		db_as.valid = true
		/* add entry to the db summary list  */
		db_list = append(db_list, db_as)
		lsid := convertUint32ToIPv4(lsaKey.LSId)
		server.logger.Info(fmt.Sprintln("negotiation: db_list AS ext append router lsid  ", lsid))
	}
	return db_list
}

/* @fn generateDbsummaryLsaList
This function will attach summary LSAs if the router is ABR
*/
func (server *OSPFServer) generateDbsummary3LsaList(self_areaId uint32) []*NbrDbSummaryData {
	db_list := []*NbrDbSummaryData{}

	lsdbKey := LsdbKey{
		AreaId: self_areaId,
	}

	area_lsa, exist := server.LsdbData.AreaLsdb[lsdbKey]
	if !exist {
		server.logger.Err(fmt.Sprintln("negotiation: Summary LSA 3 doesnt exist"))
		return nil
	}
	summary_lsdb := area_lsa.Summary3LsaMap
	selfOrigLsaEnt, _ := server.LsdbData.AreaSelfOrigLsa[lsdbKey]

	for lsaKey, _ := range summary_lsdb {
		_, exist := selfOrigLsaEnt[lsaKey]
		if exist && !server.ospfGlobalConf.isABR {
			continue // dont add self gen LSA if I am not ABR
		}
		// check if lsa instance is marked true
		db_summary := newNbrDbSummaryData()
		drlsa, ret := server.LsdbData.getSummaryLsaFromLsdb(self_areaId, lsaKey)
		if ret == LsdbEntryNotFound {
			continue
		}
		lsa_header := getLsaHeaderFromLsa(drlsa.LsaMd.LSAge, drlsa.LsaMd.Options,
			Summary3LSA, lsaKey.LSId, lsaKey.AdvRouter,
			uint32(drlsa.LsaMd.LSSequenceNum), drlsa.LsaMd.LSChecksum,
			drlsa.LsaMd.LSLen)
		/* add entry to the db summary list  */
		db_summary = append(db_summary, lsa_header)
		lsid := lsaKey.LSId
		server.logger.Info(fmt.Sprintln("negotiation: db_list summary 3 append router lsid  ", lsid))
	}
	return db_list
}

/* @fn generateDbsummaryLsaList
This function will attach summary LSAs if the router is ABR
*/
func (server *OSPFServer) generateDbsummary4LsaList(self_areaId uint32) []*NbrDbSummaryData {
	db_list := []*NbrDbSummaryData{}

	lsdbKey := LsdbKey{
		AreaId: self_areaId,
	}

	area_lsa, exist := server.LsdbData.AreaLsdb[lsdbKey]
	if !exist {
		server.logger.Err(fmt.Sprintln("negotiation: Summary LSA 4 doesnt exist"))
		return nil
	}
	summary_lsdb := area_lsa.Summary4LsaMap
	selfOrigLsaEnt, _ := server.LsdbData.AreaSelfOrigLsa[lsdbKey]

	for lsaKey, _ := range summary_lsdb {
		_, exist := selfOrigLsaEnt[lsaKey]
		if exist && !server.ospfGlobalConf.isABR {
			continue // dont add self gen LSA if I am not ABR
		}
		// check if lsa instance is marked true
		db_summary := newNbrDbSummaryData()
		drlsa, ret := server.LsdbData.getSummaryLsaFromLsdb(self_areaId, lsaKey)
		if ret == LsdbEntryNotFound {
			continue
		}
		lsa_header := getLsaHeaderFromLsa(drlsa.LsaMd.LSAge, drlsa.LsaMd.Options,
			Summary4LSA, lsaKey.LSId, lsaKey.AdvRouter,
			uint32(drlsa.LsaMd.LSSequenceNum), drlsa.LsaMd.LSChecksum,
			drlsa.LsaMd.LSLen)
		/* add entry to the db summary list  */
		db_list = append(db_list, lsa_header)
		lsid := lsaKey.LSId
		server.logger.Info(fmt.Sprintln("negotiation: db_list summary 4 append router lsid  ", lsid))
	}
	return db_list
}

func (server *OSPFServer) generateRequestList(nbrKey NbrDbdDataKey, nbrConf NbrConf, nbrDbPkt NbrDbdData) {
	/* 1) get lsa headers update in req_list */
	headers_len := len(nbrDbPkt.lsa_headers)
	server.logger.Info(fmt.Sprintln("REQ_LIST: Received lsa headers for nbr ", nbrKey,
		" no of header ", headers_len))
	req_list := NbrReqList[nbrKey]
	for i := 0; i < headers_len; i++ {
		var lsaheader ospfLSAHeader
		lsaheader = nbrDbPkt.lsa_headers[i]
		result := server.lsaAddCheck(lsaheader, nbrConf) // check lsdb
		if result {
			req := newNbrReqData()
			req = append(req, lsaheader)
		}
	}
	NbrReqList[nbrKey] = req_list
	server.logger.Info(fmt.Sprintln("REQ_LIST: updated req_list for nbr ",
		nbrKey, " req_list ", req_list))
}

func (server *OSPFServer) lsaAddCheck(lsaheader ospfLSAHeader,
	nbr NbrConf) (result bool) {

	lsa_max_age := false
	intf := server.IntfConfMap[nbr.IntfKey]
	areaId := intf.AreaId
	if lsaheader.ls_age == LSA_MAX_AGE {
		lsa_max_age = true
	}
	lsa_key := NewLsaKey()
	lsa_key.AdvRouter = lsaheader.adv_router_id
	lsa_key.LSId = lsaheader.link_state_id
	lsa_key.LSType = lsaheader.ls_type
	adv_router := lsa_key.AdvRouter
	discard := true
	discard = server.selfGenLsaCheck(*lsa_key)
	if discard {
		server.logger.Info(fmt.Sprintln("DBD: Db received self originated LSA . discard. lsa key ", *lsa_key))
		return false
	}

	switch lsaheader.ls_type {
	case RouterLSA:
		rlsa := NewRouterLsa()
		drlsa, ret := server.getRouterLsaFromLsdb(areaId, *lsa_key)
		discard, _ = server.sanityCheckRouterLsa(*rlsa, drlsa, nbr, intf, ret, lsa_max_age)

	case NetworkLSA:
		nlsa := NewNetworkLsa()
		dnlsa, ret := server.getNetworkLsaFromLsdb(areaId, *lsa_key)
		discard, _ = server.sanityCheckNetworkLsa(*lsa_key, *nlsa, dnlsa, nbr, intf, ret, lsa_max_age)

	case Summary3LSA, Summary4LSA:
		slsa := NewSummaryLsa()
		dslsa, ret := server.getSummaryLsaFromLsdb(areaId, *lsa_key)
		discard, _ = server.sanityCheckSummaryLsa(*slsa, dslsa, nbr, intf, ret, lsa_max_age)

	case ASExternalLSA:
		alsa := NewASExternalLsa()
		dalsa, ret := server.getASExternalLsaFromLsdb(areaId, *lsa_key)
		discard, _ = server.sanityCheckASExternalLsa(*alsa, dalsa, nbr, intf, intf.AreaId, ret, lsa_max_age)

	}
	if discard {
		server.logger.Info(fmt.Sprintln("DBD: LSA is not added in the request list. Adv router ", adv_router,
			" ls_type ", lsaheader.ls_type))
		return false
	}
	server.logger.Info(fmt.Sprintln("DBD: LSA append to req_list adv_router ", adv_router,
		" Lsid ", lsa_key.LSId, " lstype ", lsa_key.LSType))
	return true
}
