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
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"l3/vrrp/config"
	"l3/vrrp/debug"
	"l3/vrrp/fsm"
	"strconv"
)

type L3Intf struct {
	IfIndex int32
	// cached info for IfName is required in future
	IpAddr string
}

//type VrrpGlobalInfo struct {
type VrrpInterface struct {
	Config config.IntfCfg // Vrrp config for interface
	State  *config.State  // Vrrp state for interface
	L3     *L3Intf        // Vrrp Port Information Collected From System
	Fsm    *fsm.FSM       // Vrrp fsm information
}

func (intf *VrrpInterface) InitVrrpIntf(cfg *config.IntfCfg, l3Info *L3Intf, stCh chan *IntfState) {
	intf.Config = *cfg
	intf.L3 = *l3Info
	// Init fsm
	intf.Fsm = fsm.InitFsm(&intf.Config, l3Info, stCh)
}

func (intf *VrrpInterface) UpdateStateInfo(stInfo config.State) {
	intf.State = stInfo
}
