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

type NbrConf struct {
	IntfKey         IntfConfKey
	State           int
	InactivityTimer int32
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
	NbrLsaReTxList   []ospfLsaHeader
	NbrLsaReqList    []ospfLsaHeader
	NbrDbSummaryList []ospfLsaHeader
}

var NbrconfMap map[NbrConfKey]NbrConf

func (server *OSPFV2Server) UpdateNbrConf(nbrKey NbrConfKey, conf NbrConf, flags int32) {
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
