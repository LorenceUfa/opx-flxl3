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
	"github.com/google/gopacket/pcap"
	"net"
	"sync"
	"time"
)

type NbrData struct {
	TwoWayStatus bool
	RtrPrio      uint8
	DRtrIpAddr   uint32
	BDRtrIpAddr  uint32
	NbrIpAddr    uint32 //In case of Broadcast sorurce is NbrIpAddr
	RtrId        uint32 //In case of P2P source RtrId
}

type BackupSeenMsg struct {
	RouterId    uint32
	BDRtrIpAddr uint32
	DRtrIpAddr  uint32
}

type NbrCreateMsg struct {
	RouterId     uint32
	NbrIP        uint32
	TwoWayStatus bool
	RtrPrio      uint8
	DRtrIpAddr   uint32
	BDRtrIpAddr  uint32
	NbrKey       NbrConfKey
}

type NbrChangeMsg struct {
	RouterId     uint32
	NbrIP        uint32
	TwoWayStatus bool
	RtrPrio      uint8
	DRtrIpAddr   uint32
	BDRtrIpAddr  uint32
	NbrKey       NbrConfKey
}

type NbrConfKey struct {
	NbrIdentity         uint32
	NbrAddressLessIfIdx uint32
}

type IntfTxHandle struct {
	SendPcapHdl *pcap.Handle
	SendMutex   sync.Mutex
}

type IntfRxHandle struct {
	RecvPcapHdl        *pcap.Handle
	PktRecvCtrlCh      chan bool
	PktRecvCtrlReplyCh chan bool
}

type NetworkLSAChangeMsg struct {
	AreaId    uint32
	IntfKey   IntfConfKey
	IntfState bool
}

type NbrHelloEventMsg struct {
	IntfConfKey  IntfConfKey
	RouterId     uint32
	RtrPrio      uint8
	NbrIP        uint32
	NbrDeadTime  time.Duration
	TwoWayStatus bool
	NbrDRIpAddr  uint32
	NbrBDRIpAddr uint32
	NbrMAC       net.HardwareAddr
	NbrKey       NbrConfKey
}

type NetworkDRChangeMsg struct {
	IntfKey         IntfConfKey
	OldIntfFSMState uint8
	NewIntfFSMState uint8
}

type DeleteNbrMsg struct {
	NbrKeyList []NbrConfKey //List of Nbr Identity
}

type GenerateRouterLSAMsg struct {
	AreaId uint32
}
type NbrDownMsg struct {
	NbrKey NbrConfKey
}
type RecvdLsaMsgType uint8

const (
	LSA_ADD RecvdLsaMsgType = 0
	LSA_DEL RecvdLsaMsgType = 1
)

type RecvdLsaMsg struct {
	MsgType RecvdLsaMsgType
	LsdbKey LsdbKey
	LsaKey  LsaKey
	LsaData interface{}
}

type RecvdSelfLsaMsg struct {
	LsdbKey LsdbKey
	LsaKey  LsaKey
	LsaData interface{}
}

type LsaOp uint8

const (
	GENERATE LsaOp = 0
	FLUSH    LsaOp = 1
)

type UpdateSelfNetworkLSAMsg struct {
	Op      LsaOp
	IntfKey IntfConfKey
	NbrList []uint32
}

type LsdbToFloodLSAMsg struct {
	AreaId  uint32
	LsaKey  LsaKey
	LsaData interface{}
}

type IntfToNbrFSMChStruct struct {
	NbrHelloEventCh   chan NbrHelloEventMsg
	DeleteNbrCh       chan DeleteNbrMsg //List of Nbr Identity
	NetworkDRChangeCh chan NetworkDRChangeMsg
}

type IntfFSMToLsdbChStruct struct {
	GenerateRouterLSACh chan GenerateRouterLSAMsg
}

type NbrToIntfFSMChStruct struct {
	NbrDownMsgChMap map[IntfConfKey]chan NbrDownMsg
}

type NbrFSMToLsdbChStruct struct {
	RecvdLsaMsgCh          chan RecvdLsaMsg
	RecvdSelfLsaMsgCh      chan RecvdSelfLsaMsg
	UpdateSelfNetworkLSACh chan UpdateSelfNetworkLSAMsg
}

type LsdbToFloodChStruct struct {
	LsdbToFloodLSACh chan []LsdbToFloodLSAMsg
}

type LsdbToSPFChStruct struct {
	StartSPF chan bool
}

type SPFToLsdbChStruct struct {
	DoneSPF chan bool
}

type MessagingChStruct struct {
	IntfToNbrFSMChData  IntfToNbrFSMChStruct
	IntfFSMToLsdbChData IntfFSMToLsdbChStruct
	NbrToIntfFSMChData  NbrToIntfFSMChStruct
	NbrFSMToLsdbChData  NbrFSMToLsdbChStruct
	LsdbToFloodChData   LsdbToFloodChStruct
	LsdbToSPFChData     LsdbToSPFChStruct
	SPFToLsdbChData     SPFToLsdbChStruct
}
