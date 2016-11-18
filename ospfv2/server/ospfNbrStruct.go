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
	"encoding/binary"
	"fmt"
	"l3/ospf/config"
	"time"
)

type NbrConf struct {
	IntfKey         IntfConfKey
	State           int
	InactivityTimer int32
	RtrId           uint32
	isMaster        bool
	DDSequenceNum   uint32
	Mtu             uint32
	NbrRtrId        uint32
	NbrPriority     int32
	NbrIP           uint32
	NbrOption       uint32
	NbrDR           uint32 //mentioned by rfc.
	NbrBdr          uint32 //needed by rfc. not sure why we need it.
	NbrDeadDuration int
	NbrDeadTimer    *time.Timer
	Nbrflags        int32 //flags showing fields to update from nbrstruct
	NbrLastDbd      map[NbrConfKey]NbrDbdData
	//Nbr lists
	NbrReqList       map[NbrConfKey][]*ospfLSAHeader
	NbrDBSummaryList map[NbrConfKey][]*ospfLSAHeader
	NbrRetxList      map[NbrConfKey][]*ospfLSAHeader
}

const (
	INTF_MTU_MIN              int = 1500
	NBR_FLAG_STATE            int = 0x00000001
	NBR_FLAG_INACTIVITY_TIMER int = 0x00000002
	NBR_FLAG_IS_MASTER        int = 0x00000004
	NBR_FLAG_SEQ_NUMBER       int = 0x00000008
	NBR_FLAG_NBR_ID           int = 0x00000010
	NBR_FLAG_IP               int = 0x00000012
	NBR_FLAG_PRIORITY         int = 0x00000014
	NBR_FLAF_IP               int = 0x00000018
	NBR_FLAG_OPTIONS          int = 0x00000020
	NBR_FLAG_DR               int = 0x00000022
	NBR_FLAG_BDR              int = 0x00000024
	NBR_FLAG_DEAD_DURATION    int = 0x00000026
)

//Nbr states
//TODO move to objects
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

type NbrDbdMsg struct {
	nbrConfKey NbrConfKey
	nbrFull    bool
	nbrDbdData NbrDbdData
}

type NbrLsaReqMsg struct {
	lsa_slice []ospfLSAReq
	nbrKey    NbrConfKey
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

type NbrStruct struct {
	neighborDBDEventCh    chan NbrDbdMsg
	neighborLSAReqEventCh chan NbrLsaReqMsg
	neighborLSAUpdEventCh chan NbrLsaUpdMsg
	neighborLSAACKEventCh chan NbrLsaAckMsg
	IntfToNbrMap          map[IntfConfKey][]NbrConfKey
	nbrFSMCtrlCh          chan bool
}

func (server *OSPFV2Server) initNbrStruct() {
	server.NbrConfData.IntfToNbrMap = make(map[IntfConfKey][]NbrConfKey)
	server.NbrConfData.neighborDBDEventCh = make(chan NbrDbdMsg)
	server.NbrConfDataneighborLSAUpdEventCh = make(chan NbrLsaUpdMsg)
	server.NbrConfData.neighborLSAReqEventCh = make(chan NbrLsaReqMsg)
	server.NbrConfData.neighborLSAACKEventCh = make(chan NbrLsaAckMsg)
	server.logger.Debug("Nbr: InitNbrStruct done ")
}
