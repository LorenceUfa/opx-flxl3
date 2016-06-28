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
	_ "github.com/google/gopacket/pcap"
	"l3/ndp/config"
	"utils/dmnBase"
	"utils/logging"
)

const (
	NDP_PORT_STATE_UP   = "UP"
	NDP_PORT_STATE_DOWN = "DOWN"
	NDP_IP_STATE_UP     = "UP"
	NDP_IP_STATE_DOWN   = "DOWN"
)

type NDPServer struct {
	DmnBase *dmnBase.FSDaemon
	logger  *logging.Writer

	// System Ports information, key is IntfRef
	PhyPort map[string]config.PortInfo
	L3Port  map[string]config.IPv6IntfInfo

	ndpIntfStateSlice     []string
	ndpUpIntfStateSlice   []string
	ndpL3IntfStateSlice   []string
	ndpUpL3IntfStateSlice []string
}

const (
	NDP_SYSTEM_PORT_MAP_CAPACITY = 50
)
