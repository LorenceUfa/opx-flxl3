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

type NeighborData struct {
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

type NeighCreateMsg struct {
	RouterId     uint32
	NbrIP        uint32
	TwoWayStatus bool
	RtrPrio      uint8
	DRtrIpAddr   uint32
	BDRtrIpAddr  uint32
	NbrKey       NeighborConfKey
}

type NeighChangeMsg struct {
	RouterId     uint32
	NbrIP        uint32
	TwoWayStatus bool
	RtrPrio      uint8
	DRtrIpAddr   uint32
	BDRtrIpAddr  uint32
	NbrKey       NeighborConfKey
}

type NeighborConfKey struct {
	NbrIdentity         uint32
	NbrAddressLessIfIdx uint32
}

type NbrStateChangeMsg struct {
	NbrKey NeighborConfKey
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

type IntfToNeighMsg struct {
	IntfConfKey  IntfConfKey
	RouterId     uint32
	RtrPrio      uint8
	NeighborIP   uint32
	NbrDeadTime  time.Duration
	TwoWayStatus bool
	NbrDRIpAddr  uint32
	NbrBDRIpAddr uint32
	NbrMAC       net.HardwareAddr
	NbrKey       NeighborConfKey
}

type NetworkDRChangeMsg struct {
	IntfKey         IntfConfKey
	OldIntfFSMState uint8
	NewIntfFSMState uint8
}

type DeleteNeighborMsg struct {
	NbrKeyList []NeighborConfKey //List of Neighbor Identity
}

type IntfToNbrFSMChStruct struct {
	NeighborHelloEventCh chan IntfToNeighMsg
	DeleteNeighborCh     chan DeleteNeighborMsg //List of Neighbor Identity
	NetworkDRChangeCh    chan NetworkDRChangeMsg
}

type GenerateRouterLSAMsg struct {
	AreaId uint32
}

type IntfFSMToLsdbChStruct struct {
	GenerateRouterLSACh chan GenerateRouterLSAMsg
}
