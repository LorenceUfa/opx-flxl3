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

package server

import (
//	"l3/ospfv2/objects"
)

/* Handle neighbor events
 */
func (server *OSPFV2Server) ProcessNbrFSM() {
	for {
		select {
		//Hello packet recieved.
		case nbrData <- server.MessagingChData.IntfToNbrFSMChData.NbrHelloEventCh:
			server.logger.Debug("Nbr : Received hello event. ", nbrData.NbrIpAddr)
			server.ProcessNbrHello(nbrData)
			//DBD received

			//LSAReq received

			//LSAAck received

			//LSA received

			//Intf state change

			//Nbr dead

			//NbrFsmCtrlCh
		}
	}
}

/**** handle neighbor states ***/
func (server *OSPFV2Server) ProcessNbrHello(nbrData NbrHelloEventMsg) {
	nbrKey := NbrConfKey{
		NbrIdentity:         nbrData.NbrIpAddr,
		NbrAddressLessIfIdx: 0, // Add the index
	}
	nbrConf, valid := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Debug("Nbr: add new neighbor ", nbrData.NbrIpAddr)
		// add new neighbor
		server.CreateNewNbr(nbrData)
	}
}

func (server *OSPFV2Server) CreateNewNbr(nbrData NbrHelloEventMsg) {
	var nbrConf NbrConf
	nbrKey := NbrConfKey{
		NbrIdentity: nbrData.NbrIP,
	}
	nbrConf.NbrIP = nbrData.NbrIP
	nbrConf.NbrDR = nbrData.NbrDRIpAddr
	nbrConf.NbrBdr = nbrData.NbrBDRIpAddr
	nbrConf.IntfKey = nbrData.IntfConfKey
	nbrConf.RtrId = nbrData.RouterId
	nbrConf.NbrDeadDuration = 40
	nbrConf.InactivityTimer = time.Now()
	if nbrData.TwoWayStatus {
		nbrConf.State = NbrTwoWay
		server.ProcessNbrExtart(nbrKey, nbrConf)
		server.NbrConfMap[nbrKey] = nbrConf
	} else {
		nbrConf.State = NbrInit
		server.NbrConfMap[nbrKey] = nbrConf
	}
	newList := []NbrConfKey{}
	newList = append(newList, nbrKey)
	IntfToNbrMap[nbrData.IntfConfKey] = newList
	server.neighborDeadTimerEvent(nbrKey)
}

func (server *OSPFV2Server) ProcessNbrExtart(nbrKey NbrConfKey, nbrConf NbrConf) {
	var dbd_mdata NbrDbdData
	var flags int
	isAdjacent := server.StartAdjacency(nbrConf)
	if isAdjacent {
		nbrConf.State = NbrExchangeStart
		dbd_mdata.dd_sequence_number = uint32(time.Now().Nanosecond())
		// send dbd packets
		ticker = time.NewTicker(time.Second * 10)
		server.ConstructAndSendDbdPacket(nbrKey, true, true, true,
			INTF_OPTIONS, nbrConf.ospfNbrSeqNum, false, false, intfConf.IfMtu)
		flags = NBR_FLAG_STATE | NBR_FLAG_SEQ_NUMBER | NBR_FLAF_IS_MASTER | NBR_FLAG_INACTIVITY_TIMER
		server.logger.Debug("Nbr: Exstart seq ", dbd_mdata.dd_sequence_number)
		//Initialise all lists
		NbrReqList[nbrKey] = []*NbrReqData{}
		NbrDBSummaryList[nbrKey] = []*NbrDbSummaryData{}
		NbrRetxList[nbrKey] = []*NbrRetxData{}

	} else { // no adjacency
		server.logger.Debug("Nbr: Twoway  ", nbrKey)
		nbrConf.State = NbrTwoWay
		flags = NBR_FLAG_STATE | NBR_FLAG_INACTIVITY_TIMER
	}

	server.UpdateNbrConf(nbrKey, nbrConf, flags)

}

func (server *OSPFV2Server) ProcessNbrExchange(nbrKey NbrConfKey, dbdMsg NbrDbdData) {
}

func (server *OSPFV2Server) ProcessNbrLoading(nbrKey NbrConfKey, dbdMsg NbrDbdData) {
}

func (server *OSPFV2Server) ProcessNbrDead(nbrKey *NbrConfKey) {
	var nbr_entry_dead_func func()
	_, nbrConf := server.NbrConfMap[nbrKey]
	nbr_entry_dead_func = func() {
		server.logger.Info(fmt.Sprintln("NBRSCAN: DEAD ", nbrKey))
		server.logger.Info(fmt.Sprintln("DEAD: start processing nbr dead ", nbrKey))
		server.ResetNbrData(nbrKey, nbrConf.IntfConfKey)

		server.logger.Info(fmt.Sprintln("DEAD: end processing nbr dead ", nbrKey))

		_, exists := server.NbrConfMap[nbrConfKey]
		if exists {
			nbrConf := server.NbrConfMap[nbrConfKey]
			nbrConf.State = NbrDown
			nbrConf.InactivityTimer = time.Now()
			flags := NBR_FLAG_INACTIVITY_TIMER | NBR_FLAG_STATE
			server.ProcessNbrUpdate(nbrConf, flags)
			//should I delete or update neighbor here.
		}
	} // end of afterFunc callback
	_, exists := server.NbrConfMap[nbrConfKey]
	if exists {
		nbrConf := server.NbrConfMap[nbrConfKey]
		nbrConf.NbrDeadTimer = time.AfterFunc(nbrConf.NbrDeadTimer, nbr_entry_dead_func)
		server.NbrConfMap[nbrConfKey] = nbrConf
	}

}

func (server *OSPFV2Server) ProcessNbrUpdate(nbrKey NbrConfKey, nbrConf NbrConf, flags int32) {
	nbrConfOld, valid := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Err("Nbr: Nbr key is not valid. no updates. ", nbrKey)
		return
	}
	if flags & NBR_FLAG_STATE {
		nbrConfOld.state = nbrConf.State
	}
	if flags & NBR_FLAG_INACTIVITY_TIMER {
		nbrConfOld.InactivityTimer = nbrConf.InactivityTimer
	}
	if flags & NBR_FLAF_IS_MASTER {
		nbrConfOld.isMaster = nbrConf.isMaster
	}
	if flags & NBR_FLAG_SEQ_NUMBER {
		nbrConfOld.DDSequenceNum = nbrConf.DDSequenceNum
	}
	if flags & NBR_FLAG_NBR_ID {
		nbrConfOld.NbrId = nbrConf.NbrId
	}
	if flags & NBR_FLAG_PRIORITY {
		nbrConfOld.NbrPriority = nbrConf.NbrPriority
	}
	if flags & NBR_OPTIONS {
		nbrConfOld.NbrOptions = nbrConf.NbrOptions
	}

	if flags & NBR_DR {
		nbrConfOld.NbrDR = nbrConf.NbrDR
	}

	if flags & NBR_BDR {
		nbrConfOld.NbrBdr = nbrConf.NbrBdr
	}

	if flags & NBR_DEAD_DURATION {
		nbrConfOld.NbrDeadDuration = nbrConf.NbrDeadDuration
	}

	server.NbrConfMap[NbrKey] = nbrConf
	server.logger.Debug("Nbr: Nbr conf updated ", nbrKey)
}

/**** Utils APis *****/
func (server *OSPFV2Server) dbPacketDiscardCheck() bool {
	return false
}
func (server *OSPFV2Server) adjacencyEstablishmentCheck() bool {

	return false
}

func (server *OSPFServer) ResetNbrData(nbr NbrConfKey, intf IntfConfKey) {
	/* List of Neighbors per interface instance */
	NbrReqList[nbr] = []*NbrReqData{}
	NbrDBSummaryList[nbr] = []*NbrDbSummaryData{}
	NbrRetxList[nbr] = []*NbrRetxData{}

	nbrList, exists := IntfToNbrMap[intf]
	if !exists {
		server.logger.Info(fmt.Sprintln("DEAD: Nbr dead but intf-to-nbr map doesnt exist. ", nbr))
		return
	}
	newList := []NbrConfKey{}
	for inst := range nbrList {
		if nbrList[inst] != nbr {
			newList = append(newList, nbr)
		}
	}
	IntfToNbrMap[intf] = newList
	server.logger.Info(fmt.Sprintln("Nbr: nbrList ", newList))
}
