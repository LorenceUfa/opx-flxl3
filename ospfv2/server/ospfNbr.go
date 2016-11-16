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
	"l3/ospfv2/objects"
)

//NbrConfMap
type NbrConf struct {
	IntfKey         IntfConfKey
	State           int
	InactivityTimer int32
	RtrId           uint32
	isMaster        bool
	DDSequenceNum   uint32
	NbrId           uint32
	NbrPriority     int32
	NbrIP           uint32
	NbrOptions      uint32
	NbrDR           uint32 //mentioned by rfc.
	NbrBdr          uint32 //needed by rfc. not sure why we need it.
	NbrDeadDuration int
	NbrDeadTimer    *time.Timer
	Nbrflags        int32 //flags showing fields to update from nbrstruct
	//Nbr lists
	//	NbrLsaReTxList   []ospfLsaHeader
	//	NbrLsaReqList    []ospfLsaHeader
	//	NbrDbSummaryList []ospfLsaHeader
}

//Nbr states
type NbrState int

const (
	NbrDown          NbrState = 1
	NbrAttempt       NbrState = 2
	NbrInit          NbrState = 3
	NbrTwoWay        NbrState = 4
	NbrExchangeStart NbrState = 5
	NbrExchange      NbrState = 6
	NbrLoading       NbrState = 7
	NbrFull          NbrState = 8
)

var NbrStateList = []string{
	"Undef",
	"NbrDown",
	"NbrAttempt",
	"NbrInit",
	"NbrTwoWay",
	"NbrExchangeStart",
	"NbrExchange",
	"NbrLoading",
	"NbrFull"}

//DBD metadata
type NbrDbdData struct {
	options            uint8
	interface_mtu      uint16
	dd_sequence_number uint32
	ibit               bool
	mbit               bool
	msbit              bool
	lsa_headers        []ospfLSAHeader
}

//Lsa header
type ospfLSAHeader struct {
	ls_age          uint16
	options         uint8
	ls_type         uint8
	link_state_id   uint32
	adv_router_id   uint32
	ls_sequence_num uint32
	ls_checksum     uint16
	ls_len          uint16
}

/* LSA lists */
type NbrReqData struct {
	lsa_headers ospfLSAHeader
	valid       bool // entry is valid or not
}

func newNbrReqData() *NbrReqData {
	return &NbrReqData{}
}

type NbrDbSummaryData struct {
	lsa_headers ospfLSAHeader
	valid       bool
}

func newNbrDbSummaryData() *NbrDbSummaryData {
	return &NbrDbSummaryData{}
}

type NbrRetxData struct {
	lsa_headers ospfLSAHeader
	valid       bool
}

func newNbrRetxData() *NbrRetxData {
	return &NbrRetxData{}
}

/* neighbor lists each indexed by neighbor router id. */
var NbrReqList map[NbrConfKey][]*NbrReqData
var NbrDBSummaryList map[NbrConfKey][]*NbrDbSummaryData
var NbrRetxList map[NbrConfKey][]*NbrRetxData

//intf to nbr map
var IntfToNbrMap map[IntfConfKey][]NbrConfKey

const (
	INTF_MTU_MIN              int = 1500
	NBR_FLAG_STATE            int = 0x00000001
	NBR_FLAG_INACTIVITY_TIMER int = 0x00000002
	NBR_FLAF_IS_MASTER        int = 0x00000004
	NBR_FLAG_SEQ_NUMBER       int = 0x00000008
	NBR_FLAG_NBR_ID           int = 0x00000010
	NBR_FLAG_IP               int = 0x00000012
	NBR_FLAG_PRIORITY         int = 0x00000014
	NBR_FLAF_IP               int = 0x00000018
	NBR_OPTIONS               int = 0x00000020
	NBR_DR                    int = 0x00000022
	NBR_BDR                   int = 0x00000024
	NBR_DEAD_DURATION         int = 0x00000026
)

var NbrconfMap map[NbrConfKey]NbrConf

func (server *OSPFV2Server) UpdateNbrConf(nbrKey NbrConfKey, conf NbrConf, flags int32) {
	valid, nbrE := server.NbrConfMap[nbrKey]
	if !valid {
		server.logger.Err("Nbr : Nbr conf does not exist . Not updated ", nbrKey)
		return
	}
	if flags&NBR_FLAG_STATE == NBR_FLAG_STATE {
		nbrE.State = conf.State
	}
	if flags&NBR_FLAG_INACTIVITY_TIMER == NBR_FLAG_INACTIVITY_TIMER {
		nbrE.InactivityTimer = time.Now()
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

	server.NbrConfMap[nbrKey] = nbrE
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
		if nbrDbPkt.dd_sequence_number != nbrConf.ospfNbrSeqNum {
			server.logger.Info(fmt.Sprintln("NBREVENT:SeqNumberMismatch : Nbr is slave but dbd packet seq no doesnt match.dbd seq ",
				nbrDbPkt.dd_sequence_number, "nbr seq ", nbrConf.DDSequenceNum))
			return true
		}
	}

	return false
}

func calculateMaxLsaHeaders() (max_headers uint8) {
	rem := INTF_MTU_MIN - (OSPF_DBD_MIN_SIZE + OSPF_HEADER_SIZE)
	max_headers = uint8(rem / OSPF_LSA_HEADER_SIZE)
	return max_headers
}

func calculateMaxLsaReq() (max_req uint8) {
	rem := INTF_MTU_MIN - OSPF_HEADER_SIZE
	max_req = uint8(rem / OSPF_LSA_REQ_SIZE)
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
