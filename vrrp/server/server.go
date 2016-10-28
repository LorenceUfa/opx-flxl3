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
	"l3/vrrp/debug"
	"os"
	"os/signal"
	"syscall"
	"utils/asicdClient"
	"utils/dmnBase"
)

func (svr *VrrpServer) EventListener() {
	// Start receviing in rpc values in the channell
	for {
		select {

		case gCfg, ok := <-svr.GblCfgCh:
			if ok {
				svr.HandleGlobalConfig(gCfg)
			}
		case cfg, ok := <-svr.CfgCh:
			if ok {
				svr.HandleIntfConfig(cfg)
			}
		case l3NotifyInfo, ok := <-svr.L3IntfNotifyCh:
			if ok {
				svr.HandleIpNotification(l3NotifyInfo)
			}
		}
	}
}

func (svr *VrrpServer) GetSystemInfo() {
	// Get ports information
	svr.GetPorts()
	// Get vlans information
	svr.GetVlans()
	// Get IP Information
	svr.GetIPIntfs()
}

func (svr *VrrpServer) InitGlobalDS() {
	//svr.L2Port = make(map[int32]config.PhyPort, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V6 = make(map[int32]*V6Intf, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V4 = make(map[int32]*V4Intf, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	//svr.VlanInfo = make(map[int32]config.VlanInfo, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.Intf = make(map[KeyInfo]VrrpInterface, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V4IntfRefToIfIndex = make(map[string]int32, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	svr.V6IntfRefToIfIndex = make(map[string]int32, VRRP_GLOBAL_INFO_DEFAULT_SIZE)
	//svr.StateCh = make(chan *IntfState, VRRP_RX_BUF_CHANNEL_SIZE)
	svr.GblCfgCh = make(chan *config.GlobalConfig)
	svr.CfgCh = make(chan *config.IntfCfg, VRRP_INTF_CONFIG_CH_SIZE)
	svr.L3IntfNotifyCh = make(chan *config.BaseIpInfo)
}

func (svr *VrrpServer) DeAllocateMemory() {
	// @TODO:
}

func (svr *VrrpServer) signalHandler(sigChannel <-chan os.Signal) {
	signal := <-sigChannel
	switch signal {
	case syscall.SIGHUP:
		debug.Logger.Alert("Received SIGHUP Signal")
		//svr.VrrpCloseAllPcapHandlers()
		svr.DeAllocateMemory()
		debug.Logger.Alert("Closed all pcap's and freed memory")
		os.Exit(0)
	default:
		debug.Logger.Info("Unhandled Signal:", signal)
	}
}

func (svr *VrrpServer) OSSignalHandle() {
	sigChannel := make(chan os.Signal, 1)
	signalList := []os.Signal{syscall.SIGHUP}
	signal.Notify(sigChannel, signalList...)
	go svr.signalHandler(sigChannel)
}

func (svr *VrrpServer) VrrpStartServer() {
	svr.OSSignalHandle()
	svr.ReadDB()
	svr.InitGlobalDS()
	svr.GetSystemInfo()
	go svr.EventListener()
}

func VrrpNewServer(sPlugin asicdClient.AsicdClientIntf, dmnBase *dmnBase.FSBaseDmn) *VrrpServer {
	vrrpServer := &VrrpServer{}
	vrrpServer.SwitchPlugin = sPlugin
	vrrpServer.dmnBase = dmnBase
	/*
			svr := &VrrpServer{}
			svr.SwitchPlugin = sPlugin
			svr.dmnBase = dmnBase
		return svr
	*/
	return vrrpServer
}
