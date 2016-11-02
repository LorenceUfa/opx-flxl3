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
	"l3/vrrp/config"
	"l3/vrrp/fsm"
)

type IPIntf interface {
	Init(*config.BaseIpInfo)
	Update(*config.BaseIpInfo)
	DeInit(*config.BaseIpInfo)
	GetObjFromDb(*config.BaseIpInfo)
	SetVrrpIntfKey(*KeyInfo)
	GetVrrpIntfKey() *KeyInfo
}

type VrrpInterface struct {
	L3     *config.BaseIpInfo // Vrrp Port Information Collected From System
	Config *config.IntfCfg    // Vrrp config for interface
	Fsm    *fsm.FSM           // Vrrp fsm information
}

func (intf *VrrpInterface) InitVrrpIntf(cfg *config.IntfCfg, l3Info *config.BaseIpInfo) { //, stCh chan *IntfState) {
	intf.Config = cfg
	intf.L3 = l3Info
	// Init fsm
	intf.Fsm = fsm.InitFsm(intf.Config, l3Info)
}

func (intf *VrrpInterface) StartFsm() {
	go intf.Fsm.StartFsm()
}

func (intf *VrrpInterface) UpdateStateInfo() {

}

func (intf *VrrpInterface) StopFsm() {

}
