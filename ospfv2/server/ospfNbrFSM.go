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
		case nbrData := <-server.MessagingChData.IntfToNbrFSMChData.NbrHelloEventCh:
			server.logger.Debug("Nbr : Received hello event. ", nbrData.NbrIpAddr)
			server.ProcessNbrHello(nbrData)
			//DBD received
		case dbdData := <-server.NbrConfData.neighborDBDEventCh:
			server.logger.Debug("Nbr: Received dbd event ", dbdData)
			server.ProcessNbrDbdMsg(dbdData)

		case lsaAckData := <-server.NbrConfData.neighborLSAACKEventCh:

			//LSAReq received
		case lsaReqData := <-server.NbrConfData.neighborLSAReqEventCh:
			nbr, exists := server.NbrConfigMap[nbrLSAReqPkt.nbrKey]
			if exists && nbr.State >= config.NbrExchange {
				server.DecodeLSAReq(lsaReqData)
			}

		case lsaData := <-server.NbrConfData.neighborLSAUpdEventCh:
			nbr, exists := server.NbrConfigMap[nbrLSAUpdPkt.nbrKey]

			if exists && nbr.State >= config.NbrExchange {
				server.DecodeLSAUpd(lsaData)
			}

			//Intf state change

			//Nbr dead

			//NbrFsmCtrlCh
		}
	}
}

/**** handle neighbor states ***/
func (server *OSPFV2Server) ProcessNbrHello(nbrData NbrHelloEventMsg) {
	/*
		nbrKey := NbrConfKey{
			NbrIdentity:         nbrData.NbrIpAddr,
			NbrAddressLessIfIdx: 0,
		} */
	nbrKey := nbrData.NbrKey
	nbrConf, valid := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Debug("Nbr: add new neighbor ", nbrData.NbrIpAddr)
		// add new neighbor
		server.CreateNewNbr(nbrData)
	}
	flags := NBR_FLAG_INACTIVITY_TIMER
	server.ProcessNbrUpdate(nbrKey, nbrConf, flags)
}

func (server *OSPFV2Server) ProcessNbrDbdMsg(dbdMsg NbrDbdMsg) {
	nbrConf, exists := server.NbrConfMap[dbdMsg.nbrConfKey]
	if !exist {
		server.logger.Err("Nbr : ProcessNbrDbdMsg Nbrkey does not exist ", dbdMsg.nbrConfKey)
		return
	}
	switch nbrConf.State {
	case NbrInit, NbrExchangeStart:
		server.ProcessNbrExstart(dbdMsg.nbrConfKey, dbdMsg.nbrDbdData)
	case NbrExchange:
		server.ProcessNbrExchange(dbdMsg.nbrConfKey, dbdMsg.nbrDbdData)
	case NbrLoading:
		sever.ProcessNbrLoading(dbdMsg.nbrConfKey, dbdMsg.nbrDbdData)
	case NbrFuLL:
		server.logger.Err("Nbr: Received dbd packet when nbr is full . Restart FSM", dbdMsg.nbrConfKey)
	case NbrDown:
		server.logger.Warning("Nbr: Nbr is down state. Dont process dbd ", dbdMsg.nbrConfKey)
	case NbrTwoWay:
		server.logger.Warning("Nbr: Nbr is two way state.Dont process dbd ", dbdMsg.nbrConfKey)
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
		server.ProcessNbrFsmStart(nbrKey, nbrConf)
		server.NbrConfMap[nbrKey] = nbrConf
	} else {
		nbrConf.State = NbrInit
		server.NbrConfMap[nbrKey] = nbrConf
	}
	newList := []NbrConfKey{}
	newList = append(newList, nbrKey)
	IntfToNbrMap[nbrData.IntfConfKey] = newList
	nbrConf.NbrLastDbd = make(map[nbrKey]NbrDbdData)
	server.neighborDeadTimerEvent(nbrKey)
}

func (server *OSPFV2Server) ProcessNbrFsmStart(nbrKey NbrConfKey, nbrConf NbrConf) {
	var dbd_mdata NbrDbdData
	var flags int
	isAdjacent := server.AdjacencyCheck(nbrKey)
	if isAdjacent {
		nbrConf.State = NbrExchangeStart
		dbd_mdata.dd_sequence_number = uint32(time.Now().Nanosecond())
		// send dbd packets
		ticker = time.NewTicker(time.Second * 10)
		server.ConstructDbdMdata(nbrKey, true, true, true,
			INTF_OPTIONS, nbrConf.DDSequenceNum, false, false, intfConf.IfMtu)
		server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
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

func (server *OSPFV2Server) ProcessNbrExstart(nbrKey NbrConfKey, nbrDbPkt NbrDbdData) {
	valid, nbrConf := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Err("Nbr : Extart Neighbor does not exist ", nbrKey)
		return
	}

	var dbd_mdata NbrDbdData
	last_exchange := true
	var isAdjacent bool
	var negotiationDone bool
	isAdjacent = server.AdjacencyCheck(nbrKey)
	if isAdjacent || nbrConf.State == NbrExchangeStart {
		// change nbr state
		nbrConf.State = NbrExchangeStart
		// decide master slave relation
		if nbrConf.NbrRtrId > binary.BigEndian.Uint32(server.globalData.RouterId) {
			nbrConf.isMaster = true
		} else {
			nbrConf.isMaster = false
		}
		/* The initialize(I), more (M) and master(MS) bits are set,
		   the contents of the packet are empty, and the neighbor's
		   Router ID is larger than the router's own.  In this case
		   the router is now Slave.  Set the master/slave bit to
		   slave, and set the neighbor data structure's DD sequence
		   number to that specified by the master.
		*/
		server.logger.Debug(fmt.Sprintln("NBRDBD: nbr ip ", nbrKey.NbrIP,
			" my router id ", binary.BigEndian.Uint32(server.globalData.RouterId),
			" nbr_seq ", nbrConf.DDSequenceNum, "dbd_seq no ", nbrDbPkt.dd_sequence_number))
		if nbrDbPkt.ibit && nbrDbPkt.mbit && nbrDbPkt.msbit &&
			nbrConf.NbrRtrId > binary.BigEndian.Uint32(server.globalData.RouterId) {
			server.logger.Debug(fmt.Sprintln("DBD: (ExStart/slave) SLAVE = self,  MASTER = ", nbrKey.NbrIP))
			nbrConf.isMaster = true
			server.logger.Debug("NBREVENT: Negotiation done..")
			negotiationDone = true
			nbrConf.State = config.NbrExchange
		}
		if nbrDbPkt.msbit && nbrConf.NbrRtrId > binary.BigEndian.Uint32(server.globalData.RouterId) {
			server.logger.Debug(fmt.Sprintln("DBD: (ExStart/slave) SLAVE = self,  MASTER = ", nbrKey.NbrIP))
			nbrConf.isMaster = true
			server.logger.Debug("NBREVENT: Negotiation done..")
			negotiationDone = true
			nbrConf.State = config.NbrExchange
		}

		/*   The initialize(I) and master(MS) bits are off, the
		     packet's DD sequence number equals the neighbor data
		     structure's DD sequence number (indicating
		     acknowledgment) and the neighbor's Router ID is smaller
		     than the router's own.  In this case the router is
		     Master.
		*/
		if nbrDbPkt.ibit == false && nbrDbPkt.msbit == false &&
			nbrDbPkt.dd_sequence_number == nbrConf.DDSequenceNum &&
			nbrConf.NbrRtrId < binary.BigEndian.Uint32(server.globalData.RouterId) {
			nbrConf.isMaster = false
			server.logger.Debug(fmt.Sprintln("DBD:(ExStart) SLAVE = ", nbrKey.NbrIP, "MASTER = SELF"))
			server.logger.Debug("NBREVENT: Negotiation done..")
			negotiationDone = true
			nbrConf.State = config.NbrExchange
		}

	} else {
		nbrConf.State = config.NbrTwoWay
	}

	var lsa_attach uint8
	if negotiationDone {
		//server.logger.Debug(fmt.Sprintln("DBD: (Exstart) lsa_headers = ", len(nbrDbPkt.lsa_headers)))
		server.generateDbSummaryList(nbrKey)

		if nbrConf.isMaster != true { // i am the master
			dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, false, true, true,
				nbrDbPkt.options, nbrDbPkt.dd_sequence_number+1, true, false)
			server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
		} else {
			// send acknowledgement DBD with I and MS bit false , mbit = 1
			dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, false, true, false,
				nbrDbPkt.options, nbrDbPkt.dd_sequence_number, true, false)
			server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
			dbd_mdata.dd_sequence_number++
		}

		if last_exchange {
			nbrConf.nbrEvent = config.NbrExchangeDone
		}
		server.generateRequestList(nbrKey, nbrConf, nbrDbPkt)

	} else { // negotiation not done
		nbrConf.State = NbrExchangeStart
		if nbrConf.isMaster &&
			nbrConf.NbrRtrId > binary.BigEndian.Uint32(server.globalData.RouterId) {
			dbd_mdata.dd_sequence_number = nbrDbPkt.dd_sequence_number
			dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, true, true, true,
				nbrDbPkt.options, nbrDbPkt.dd_sequence_number, false, false)
			server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
			dbd_mdata.dd_sequence_number++
		} else {
			//start with new seq number
			dbd_mdata.dd_sequence_number = uint32(time.Now().Nanosecond()) //nbrConf.DDSequenceNum
			dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, true, true, true,
				nbrDbPkt.options, nbrDbPkt.dd_sequence_number, false, false)
			server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
		}
	}
	nbrConf.dd_sequence_number = nbrDbPkt.dd_sequence_number
	nbrConf.options = nbrDbPkt.options
	flags := NBR_FLAG_SEQ_NUMBER | NBR_FLAG_OPTIONS | NBR_FLAG_INACTIVITY_TIMER | NBR_FLAG_IS_MASTER
	server.ProcessNbrUpdate(nbrKey, nbrConf, flags)

}

func (server *OSPFV2Server) ProcessNbrExchange(nbrKey NbrConf, nbrDbPkt NbrDbdData) {
	var last_exachange bool
	valid, nbrConf := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Err("Nbr : Exchange Neighbor does not exist ", nbrKey)
		return
	}

	var nbrState NbrState
	isDiscard := server.exchangePacketDiscardCheck(nbrConf, nbrDbPkt)
	if isDiscard {
		server.logger.Debug(fmt.Sprintln("NBRDBD: (Exchange)Discard packet. nbr", nbrKey.NbrIP,
			" nbr state ", nbrConf.State))

		nbrState = NbrExchangeStart
		nbrConf.State = NbrExchangeStart
		server.ProcessNbrExstart(nbrKey, nbrDbPkt)

		return
	} else { // process exchange state
		/* 2) Add lsa_headers to db packet from db_summary list */

		if nbrConf.isMaster != true { // i am master
			/* Send the DBD only if packet has mbit =1 or event != NbrExchangeDone
			          send DBD with seq num + 1 , ibit = 0 ,  ms = 1
			   * if this is the last DBD for LSA description set mbit = 0
			*/
			server.logger.Debug(fmt.Sprintln("DBD:(master/Exchange) nbr_event ", nbrConf.nbrEvent, " mbit ", nbrDbPkt.mbit))
			if nbrDbPkt.dd_sequence_number == nbrConf.DDSequenceNum &&
				(nbrConf.nbrEvent != config.NbrExchangeDone ||
					nbrDbPkt.mbit) {
				server.logger.Debug(fmt.Sprintln("DBD: (master/Exchange) Send next packet in the exchange  to nbr ", nbrKey.NbrIP))
				dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, false, false, true,
					nbrDbPkt.options, nbrDbPkt.dd_sequence_number+1, true, false, intfConf.IfMtu)
				server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
				NbrLastDbd[nbrKey] = dbd_mdata
			}

			// Genrate request list
			server.generateRequestList(nbrKey, nbrConf, nbrDbPkt)
			server.logger.Debug(fmt.Sprintln("DBD:(Exchange) Total elements in req_list ", len(ospfNeighborRequest_list[nbrKey])))

		} else { // i am slave
			/* send acknowledgement DBD with I and MS bit false and mbit same as
			   rx packet
			    if mbit is 0 && last_exchange == true generate NbrExchangeDone*/
			if nbrDbPkt.dd_sequence_number == nbrConf.DDSequenceNum {
				server.logger.Debug(fmt.Sprintln("DBD: (slave/Exchange) Send next packet in the exchange  to nbr ", nbrKey.NbrIP))
				server.generateRequestList(nbrKey, nbrConf, nbrDbPkt)
				dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, false, nbrDbPkt.mbit, false,
					nbrDbPkt.options, nbrDbPkt.dd_sequence_number, true, false, intfConf.IfMtu)
				server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
				NbrLastDbd[nbrKey] = dbd_mdata
				dbd_mdata.dd_sequence_number++
			} else {
				server.logger.Debug(fmt.Sprintln("DBD: (slave/exchange) Duplicated dbd.  . dbd_seq , nbr_seq_num ",
					nbrDbPkt.dd_sequence_number, nbrConf.DDSequenceNum))
				if !nbrDbPkt.msbit && !nbrDbPkt.ibit {
					// the last exchange packet so we need not send duplicate response
					last_exchange = true
				}
				// send old ACK
				data := newDbdMsg(nbrKey, NbrLastDbd[nbrKey])
				dbd_mdata = NbrLastDbd[nbrKey]
				server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
			}
		}
		if !nbrDbPkt.mbit && last_exchange {
			nbrState = NbrLoading
			nbrConf.ospfNbrLsaReqIndex = server.BuildAndSendLSAReq(nbrKey, nbrConf)
			server.logger.Debug(fmt.Sprintln("DBD: Loading , nbr ", nbrKey.NbrIP))
		}
	}
	nbrConf.dd_sequence_number = nbrDbPkt.dd_sequence_number
	flags := NBR_FLAG_SEQ_NUMBER | NBR_FLAG_OPTIONS | NBR_FLAG_INACTIVITY_TIMER | NBR_FLAG_STATE
	server.ProcessNbrUpdate(nbrKey, nbrConf, flags)

}

func (server *OSPFV2Server) ProcessNbrLoading(nbrKey NbrConf, nbrDbPkt NbrDbdData) {
	var seq_num uint32
	server.logger.Debug(fmt.Sprintln("DBD: Loading . Nbr ", nbrKey.NbrIP))
	isDiscard := server.exchangePacketDiscardCheck(nbrConf, nbrDbPkt)
	isDuplicate := server.verifyDuplicatePacket(nbrConf, nbrDbPkt)

	if isDiscard {
		server.logger.Debug(fmt.Sprintln("NBRDBD:Loading  Discard packet. nbr", nbrKey.NbrIP,
			" nbr state ", nbrConf.State))
		//update neighbor to exchange start state and send dbd

		nbrConf.State = NbrExchangeStart
		nbrConf.isMaster = false
		dbd_mdata, last_exchange = server.ConstructDbdMdata(nbrKey, true, true, true,
			nbrDbPkt.options, nbrConf.DDSequenceNum+1, false, false, intfConf.IfMtu)
		server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
		seq_num = dbd_mdata.dd_sequence_number
	} else if !isDuplicate {
		/*
		   slave - Send the old dbd packet.
		       master - discard
		*/
		if nbrConf.isMaster {
			dbd_mdata, _ := server.ConstructDbdMdata(nbrKey, false, nbrDbPkt.mbit, false,
				nbrDbPkt.options, nbrDbPkt.dd_sequence_number, false, false, intfConf.IfMtu)
			server.BuildAndSendDdBDPkt(nbrConf, dbd_mdata)
			seq_num = dbd_mdata.dd_sequence_number + 1
		}
		seq_num = NbrLastDbd[nbrKey].dd_sequence_number
		nbrConf.State = NbrLoading
	} else {
		seq_num = NbrLastDbd[nbrKey].dd_sequence_number
		nbrConf.State = NbrLoading
	}
	nbrConf.dd_sequence_number = nbrDbPkt.dd_sequence_number
	flags := NBR_FLAG_SEQ_NUMBER | NBR_FLAG_OPTIONS | NBR_FLAG_INACTIVITY_TIMER | NBR_FLAG_STATE
	server.ProcessNbrUpdate(nbrKey, nbrConf, flags)
}

func (server *OSPFV2Server) ProcessNbrFull(nbrKey NbrConfKey) {
	nbrConf, valid := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Err("Nbr: Full event , nbr key does not exist ", nbrKey)
	}
	nbrConf.State = NbrFuLL
	server.UpdateIntfToNbrMap(nbrKey)
	//TODO send message to generate networkLSA
	flags := NBR_FLAG_STATE
	server.UpdateNbrConf(nbrConf)
	server.logger.Debug("Nbr: Nbr full event ", nbrKey)
}

func (server *OSPFV2Server) ProcessNbrDead(nbrKey *NbrConfKey) {
	var nbr_entry_dead_func func()
	nbrConf, _ := server.NbrConfMap[nbrKey]
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
func (server *OSPFV2Server) AdjacencyCheck(nbrKey NbrConfKey) bool {
	nbrConf, valid := server.NbrConfMap[nbrKey]
	if !valid {
		sever.logger.Err("Nbr : Nbr does not exist . No adjacency.", nbrKey)
		return false
	}
	/*
			 o   The underlying network type is point-to-point

		        o   The underlying network type is Point-to-MultiPoint

		        o   The underlying network type is virtual link

		        o   The router itself is the Designated Router

		        o   The router itself is the Backup Designated Router

		        o   The neighboring router is the Designated Router

		        o   The neighboring router is the Backup Designated Router
	*/
	return true
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
