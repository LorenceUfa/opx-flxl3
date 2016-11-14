//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
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
//"asicd/asicdCommonDefs"
//"errors"
//"fmt"
//"ribd"
//"strconv"
)

type DestType uint8

const (
	Network         DestType = 0
	InternalRouter  DestType = 1
	ASBdrRouter     DestType = 2
	AreaBdrRouter   DestType = 3
	ASAreaBdrRouter DestType = 4
)

type PathType int

const (
	/* Decreasing order of Precedence */
	IntraArea PathType = 4
	InterArea PathType = 3
	Type1Ext  PathType = 2
	Type2Ext  PathType = 1
)

type IfData struct {
	IfIpAddr uint32
	IfIdx    uint32
}

type NbrIP uint32

type NextHop struct {
	IfIPAddr  uint32
	IfIdx     uint32
	NextHopIP uint32
	AdvRtr    uint32 // Nbr Router Id
}

type AreaIdKey struct {
	AreaId uint32
}

type RoutingTblEntryKey struct {
	DestId   uint32   // IP address(Network Type) RouterID(Router Type)
	AddrMask uint32   // Only For Network Type
	DestType DestType // true: Network, false: Router
}

type AreaRoutingTbl struct {
	RoutingTblMap map[RoutingTblEntryKey]RoutingTblEntry
}

type RoutingTblEntry struct {
	OptCapabilities uint8    // Optional Capabilities
	PathType        PathType // Path Type
	Cost            uint16
	Type2Cost       uint16
	LSOrigin        LsaKey
	NumOfPaths      int
	NextHops        map[NextHop]bool // Next Hop
}

type GlobalRoutingTblEntry struct {
	AreaId        uint32 // Area
	RoutingTblEnt RoutingTblEntry
}

func (server *OSPFV2Server) UpdateRoutingTblWithStub(areaId uint32, vKey VertexKey, tVertex TreeVertex, parent TreeVertex, parentKey VertexKey, rootVKey VertexKey) {
	//TODO
}

func (server *OSPFV2Server) UpdateRoutingTblForRouter(areaIdKey AreaIdKey, vKey VertexKey, tVertex TreeVertex, rootVKey VertexKey) {
	//TODO
}

func (server *OSPFV2Server) UpdateRoutingTblForTNetwork(areaIdKey AreaIdKey, vKey VertexKey, tVertex TreeVertex, rootVKey VertexKey) {
	//TODO
}

func (server *OSPFV2Server) InstallRoutingTbl() {
	//TODO
}
