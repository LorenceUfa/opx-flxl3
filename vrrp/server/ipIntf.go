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
)

//type VrrpGlobalInfo struct {
type VrrpInterface struct {
	// Vrrp config for interface
	Cfg config.IntfCfg
	// VRRP MAC aka VMAC
	VirtualRouterMACAddress string
	// The initial value is the same as Advertisement_Interval.
	//MasterAdverInterval int32
	// (((256 - priority) * Master_Adver_Interval) / 256)
	SkewTime int32
	// (3 * Master_Adver_Interval) + Skew_time
	MasterDownValue int32
	MasterDownTimer *time.Timer
	MasterDownLock  *sync.RWMutex
	// Advertisement Timer
	AdverTimer *time.Timer
	// IpAddr which needs to be used if no Virtual Ip is specified
	IpAddr string
	// cached info for IfName is required in future
	IfName string
	// Pcap Handler for receiving packets
	pHandle *pcap.Handle
	// Pcap Handler lock to write data one routine at a time
	PcapHdlLock *sync.RWMutex
	// State Name
	StateName string
	// Lock to read current state of vrrp object
	StateNameLock *sync.RWMutex
	// Vrrp State Lock for each IfIndex + VRID
	StateInfo     VrrpGlobalStateInfo
	StateInfoLock *sync.RWMutex
	/*
		IntfConfig vrrpd.VrrpIntf
		// VRRP MAC aka VMAC
		VirtualRouterMACAddress string
		// The initial value is the same as Advertisement_Interval.
		MasterAdverInterval int32
		// (((256 - priority) * Master_Adver_Interval) / 256)
		SkewTime int32
		// (3 * Master_Adver_Interval) + Skew_time
		MasterDownValue int32
		MasterDownTimer *time.Timer
		MasterDownLock  *sync.RWMutex
		// Advertisement Timer
		AdverTimer *time.Timer
		// IfIndex IpAddr which needs to be used if no Virtual Ip is specified
		IpAddr string
		// cached info for IfName is required in future
		IfName string
		// Pcap Handler for receiving packets
		pHandle *pcap.Handle
		// Pcap Handler lock to write data one routine at a time
		PcapHdlLock *sync.RWMutex
		// State Name
		StateName string
		// Lock to read current state of vrrp object
		StateNameLock *sync.RWMutex
		// Vrrp State Lock for each IfIndex + VRID
		StateInfo     VrrpGlobalStateInfo
		StateInfoLock *sync.RWMutex
	*/
}

func (intf *IpIntf) InitIpIntf(obj interface{}) {

}
