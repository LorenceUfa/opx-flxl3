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
	"fmt"
	"l3/ospfv2/objects"
)

func newDbdMsg(key NbrConfKey, dbd_data NbrDbdData) NbrDbdMsg {
	dbdNbrMsg := NbrDbdMsg{
		nbrConfKey: key,
		nbrDbdData: dbd_data,
	}
	return dbdNbrMsg
}

func (server *OSPFV2Server) UpdateNbrConf(nbrKey NbrConfKey, conf NbrConf, flags int) {
	nbrE, valid := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Err("Nbr : Nbr conf does not exist . Not updated ", nbrKey)
		return
	}
	if flags&NBR_FLAG_STATE == NBR_FLAG_STATE {
		nbrE.State = conf.State
	}
	if flags&NBR_FLAG_DEAD_TIMER == NBR_FLAG_DEAD_TIMER {
		server.logger.Debug("Nbr : Nbr inactivity reset ", nbrKey)
	}
	if flags&NBR_FLAG_SEQ_NUMBER == NBR_FLAG_SEQ_NUMBER {
		nbrE.DDSequenceNum = conf.DDSequenceNum
	}

	if flags&NBR_FLAG_IS_MASTER == NBR_FLAG_IS_MASTER {
		nbrE.isMaster = conf.isMaster
	}
	if flags&NBR_FLAG_PRIORITY == NBR_FLAG_PRIORITY {
		nbrE.NbrPriority = conf.NbrPriority
	}
	if flags&NBR_FLAG_OPTION == NBR_FLAG_STATE {
		nbrE.NbrOption = conf.NbrOption
	}
	server.NbrConfMap[nbrKey] = nbrE
}

func (server *OSPFV2Server) UpdateIntfToNbrMap(nbrKey NbrConfKey) {
	var newList []NbrConfKey
	nbrConf := server.NbrConfMap[nbrKey]
	_, exists := server.NbrConfData.IntfToNbrMap[nbrConf.IntfKey]
	if !exists {
		newList = []NbrConfKey{}
	} else {
		newList = server.NbrConfData.IntfToNbrMap[nbrConf.IntfKey]
	}
	newList = append(newList, nbrKey)
	server.NbrConfData.IntfToNbrMap[nbrConf.IntfKey] = newList
	server.logger.Debug("Nbr : Intf to nbr list updated ", newList)
}

func (server *OSPFV2Server) NbrDbPacketDiscardCheck(nbrDbPkt NbrDbdData, nbrConf NbrConf) bool {
	if nbrDbPkt.msbit != nbrConf.isMaster {
		server.logger.Info(fmt.Sprintln("NBREVENT: SeqNumberMismatch. Nbr should be master  dbdmsbit ", nbrDbPkt.msbit,
			" isMaster ", nbrConf.isMaster))
		return true
	}

	if nbrDbPkt.ibit == true {
		server.logger.Info("NBREVENT:SeqNumberMismatch . Nbr ibit is true ", nbrConf.NbrIP)
		return true
	}

	if nbrConf.isMaster {
		if nbrDbPkt.dd_sequence_number != nbrConf.DDSequenceNum {
			if nbrDbPkt.dd_sequence_number+1 == nbrConf.DDSequenceNum {
				server.logger.Debug(fmt.Sprintln("Duplicate: This is db duplicate packet. Ignore."))
				return false
			}
			server.logger.Info(fmt.Sprintln("NBREVENT:SeqNumberMismatch : Nbr is master but dbd packet seq no doesnt match. dbd seq ",
				nbrDbPkt.dd_sequence_number, "nbr seq ", nbrConf.DDSequenceNum))
			return true
		}
	} else {
		if nbrDbPkt.dd_sequence_number != nbrConf.DDSequenceNum {
			server.logger.Info(fmt.Sprintln("NBREVENT:SeqNumberMismatch : Nbr is slave but dbd packet seq no doesnt match.dbd seq ",
				nbrDbPkt.dd_sequence_number, "nbr seq ", nbrConf.DDSequenceNum))
			return true
		}
	}

	return false
}

func (server *OSPFV2Server) verifyDuplicatePacket(nbrConf NbrConf, nbrDbPkt NbrDbdData) (isDup bool) {
	if nbrConf.isMaster {
		if nbrDbPkt.dd_sequence_number+1 == nbrConf.DDSequenceNum {
			isDup = true
			server.logger.Info(fmt.Sprintln("NBREVENT: Duplicate packet Dont do anything. dbdseq ",
				nbrDbPkt.dd_sequence_number, " nbrseq ", nbrConf.DDSequenceNum))
			return
		}
	}
	isDup = false
	return
}

func calculateMaxLsaHeaders() (max_headers uint8) {
	rem := INTF_MTU_MIN - (OSPF_DBD_MIN_SIZE + OSPF_HEADER_SIZE)
	max_headers = uint8(rem / OSPF_LSA_HEADER_SIZE)
	return max_headers
}

func calculateMaxLsaReq() (max_req int) {
	rem := INTF_MTU_MIN - OSPF_HEADER_SIZE
	max_req = rem / OSPF_LSA_REQ_SIZE
	return max_req
}

func (server *OSPFV2Server) getNbrState(ipAddr, addressLessIfIdx uint32) (*objects.Ospfv2NbrState, error) {
	var retObj objects.Ospfv2NbrState
	return &retObj, nil
}

func (server *OSPFV2Server) getBulkNbrState(fromIdx, cnt int) (*objects.Ospfv2NbrStateGetInfo, error) {
	var retObj objects.Ospfv2NbrStateGetInfo
	return &retObj, nil
}
