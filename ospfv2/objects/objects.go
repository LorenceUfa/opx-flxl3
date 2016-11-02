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

const (
	AUTH_TYPE_NONE_STR            string = "none"
	AUTH_TYPE_SIMPLE_PASSWORD_STR string = "simplepassword"
	AUTH_TYPE_MD5_STR             string = "md5"
)

const (
	AUTH_TYPE_NONE            uint8 = 0
	AUTH_TYPE_SIMPLE_PASSWORD uint8 = 1
	AUTH_TYPE_MD5             uint8 = 2
)

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

const (
	GLOBAL_ADMIN_STATE_UP   bool = true
	GLOBAL_ADMIN_STATE_DOWN bool = false
)

const (
	GLOBAL_ADMIN_STATE_UP_STR   string = "up"
	GLOBAL_ADMIN_STATE_DOWN_STR string = "down"
)

type Ospfv2Global struct {
	Vrf                string
	RouterId           uint32
	AdminState         bool
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

const (
	INTF_ADMIN_STATE_DOWN bool = false
	INTF_ADMIN_STATE_UP   bool = true
)

const (
	INTF_ADMIN_STATE_DOWN_STR string = "down"
	INTF_ADMIN_STATE_UP_STR   string = "up"
)

const (
	INTF_TYPE_POINT2POINT_STR string = "pointtopoint"
	INTF_TYPE_BROADCAST_STR   string = "broadcast"
)

const (
	INTF_TYPE_POINT2POINT uint8 = 0
	INTF_TYPE_BROADCAST   uint8 = 1
)

const (
	INTF_FSM_STATE_OTHER_DR uint8 = 0
	INTF_FSM_STATE_DR       uint8 = 1
	INTF_FSM_STATE_BDR      uint8 = 2
	INTF_FSM_STATE_LOOPBACK uint8 = 3
	INTF_FSM_STATE_DOWN     uint8 = 4
	INTF_FSM_STATE_WAITING  uint8 = 5
	INTF_FSM_STATE_P2P      uint8 = 6
)

const (
	INTF_FSM_STATE_OTHER_DR_STR string = "other-dr"
	INTF_FSM_STATE_DR_STR       string = "dr"
	INTF_FSM_STATE_BDR_STR      string = "bdr"
	INTF_FSM_STATE_LOOPBACK_STR string = "loopback"
	INTF_FSM_STATE_DOWN_STR     string = "down"
	INTF_FSM_STATE_WAITING_STR  string = "waiting"
	INTF_FSM_STATE_P2P_STR      string = "point-to-point"
)

type Ospfv2Intf struct {
	IpAddress        uint32
	AddressLessIfIdx uint32
	AdminState       bool
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
	State                    uint8
	DesignatedRouter         uint32
	DesignatedRouterId       uint32
	BackupDesignatedRouter   uint32
	BackupDesignatedRouterId uint32
	NumNbrs                  uint32
}
type Ospfv2IntfStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*Ospfv2IntfState
}

const (
	ROUTER_LSA     uint8 = 1
	NETWORK_LSA    uint8 = 2
	SUMMARY3_LSA   uint8 = 3
	SUMMARY4_LSA   uint8 = 4
	ASExternal_LSA uint8 = 5
)

const (
	ROUTER_LSA_STR     string = "router"
	NETWORK_LSA_STR    string = "network"
	SUMMARY3_LSA_STR   string = "summary3"
	SUMMARY4_LSA_STR   string = "summary4"
	ASExternal_LSA_STR string = "asexternal"
)

type Ospfv2LsdbState struct {
	LSType        uint8
	LSId          uint32
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

const (
	NBR_STATE_ONE_WAY_STR  string = "oneway"
	NBR_STATE_TWO_WAY_STR  string = "twoway"
	NBR_STATE_INIT_STR     string = "init"
	NBR_STATE_EXSTART_STR  string = "exstart"
	NBR_STATE_EXCHANGE_STR string = "exchange"
	NBR_STATE_LOADING_STR  string = "loading"
	NBR_STATE_ATTEMPT_STR  string = "attempt"
	NBR_STATE_DOWN_STR     string = "down"
	NBR_STATE_FULL_STR     string = "full"
)

const (
	NBR_STATE_ONE_WAY  uint8 = 0
	NBR_STATE_TWO_WAY  uint8 = 1
	NBR_STATE_INIT     uint8 = 2
	NBR_STATE_EXSTART  uint8 = 3
	NBR_STATE_EXCHANGE uint8 = 4
	NBR_STATE_LOADING  uint8 = 5
	NBR_STATE_ATTEMPT  uint8 = 6
	NBR_STATE_DOWN     uint8 = 7
	NBR_STATE_FULL     uint8 = 8
)

type Ospfv2NbrState struct {
	IpAddr           uint32
	AddressLessIfIdx uint32
	RtrId            uint32
	Options          int32
	State            uint8
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
