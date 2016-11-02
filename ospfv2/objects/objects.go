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

package objects

type Ospfv2Area struct {
	AreaId   uint32
	AuthType uint8
}

type Ospfv2AreaState struct {
	AreaId           uint32
	NumSpfRuns       uint32
	NumBdrRtr        uint32
	NumAsBdrRtr      uint32
	NumRouterLsa     uint32
	NumNetworkLsa    uint32
	NumSummary3Lsa   uint32
	NumSummary4Lsa   uint32
	NumASExternalLsa uint32
	NumIntfs         uint32
	NumNbr           uint32
}

type Ospfv2AreaStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2AreaState
}

type Ospfv2LsdbState struct {
	Type          uint8
	LsId          uint32
	AreaId        uint32
	AdvRouterId   uint32
	SequenceNum   uint32
	Age           uint16
	Checksum      uint16
	Advertisement string
}

type Ospfv2LsdbStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2LsdbState
}

type Ospfv2Intf struct {
	IpAddress        uint32
	AddressLessIfIdx uint32
	AdminState       uint8
	AreaId           uint32
	Type             uint8
	RtrPriority      uint8
	TransitDelay     uint16
	RetransInterval  uint16
	HelloInterval    uint16
	RtrDeadInterval  uint32
	MetricValue      uint16
}

type Ospfv2IntfState struct {
	IpAddress                uint32
	AddressLessIfIdx         uint32
	State                    uint32
	DesignatedRouter         uint32
	DesignatedRouterId       uint32
	BackupDesignatedRouter   uint32
	BackupDesignatedRouterId uint32
	Events                   uint32
	LsaCount                 uint32
	NumNbr                   uint32
}
type Ospfv2IntfStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2IntfState
}

type Ospfv2NbrState struct {
	IpAddr           uint32
	AddressLessIfIdx uint32
	RtrId            uint32
	Options          int32
	State            string
	Events           uint32
}

type Ospfv2NbrStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2NbrState
}

type Ospfv2LsaKey struct {
	LsType    uint8
	LSId      uint32
	AdvRouter uint32
}

type Ospfv2NextHop struct {
	IntfIPAddr    uint32
	IntfIfIdx     uint32
	NextHopIPAddr uint32
	AdvRtrId      uint32
}

type Ospfv2RouteState struct {
	DestId          uint32
	AddrMask        uint32
	DestType        uint8
	OptCapabilities int32
	AreaId          uint32
	PathType        uint8
	Cost            uint32
	Type2Cost       uint32
	NumOfPaths      uint16
	NextHops        []Ospfv2NextHop
	LSOrigin        Ospfv2LsaKey
}

type Ospfv2RouteStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2RouteState
}

type Ospfv2Global struct {
	Vrf                string
	RouterId           uint32
	AdminState         uint8
	ASBdrRtrStatus     bool
	ReferenceBandwidth uint32
}

type Ospfv2GlobalState struct {
	Vrf              string
	RouterId         uint32
	AreaBdrRtrStatus bool
}

type Ospfv2GlobalStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2GlobalState
}
