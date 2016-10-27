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
	"net"
	"utils/commonDefs"
)

func (svr *VrrpServer) GetPorts() {
	debug.Logger.Info("Get Port State List")
	portsInfo, err := svr.SwitchPlugin.GetAllPortState()
	if err != nil {
		debug.Logger.Err("Failed to get all ports from system, ERROR:", err)
		return
	}
	for _, obj := range portsInfo {
		var empty struct{}
		port := config.PhyPort{
			IntfRef:   obj.IntfRef,
			IfIndex:   obj.IfIndex,
			OperState: obj.OperState,
		}
		pObj, err := svr.SwitchPlugin.GetPort(obj.IntfRef)
		if err != nil {
			debug.Logger.Err("Getting mac address for", obj.IntfRef, "failed, error:", err)
		} else {
			port.MacAddr = pObj.MacAddr
		}
		l2Port := svr.L2Port[port.IfIndex]
		l2Port.Info = port
		svr.L2Port[port.IfIndex] = l2Port
		//svr.SwitchMacMapEntries[port.MacAddr] = empty
		//svr.SwitchMac = port.MacAddr // @HACK.... need better solution
	}

	debug.Logger.Info("Done with Port State list")
	return
}

func (svr *VrrpServer) GetVlans() {

}

func (svr *VrrpServer) getIPv4Intfs() {
	ipv4Info, err := svr.SwitchPlugin.GetAllIPv4IntfState()
	if err != nil {
		debug.Logger.Err("Failed to get all IPv4 interfaces, err:", err)
		return
	}

	for _, obj := range ipv4Info {
		// do not care for loopback interface
		if svr.SwitchPlugin.IsLoopbackType(obj.IfIndex) {
			continue
		}
		v4Info, _ := svr.V4[obj.IfIndex]
		ipInfo := v4Info.Cfg.Info
		//if !exists {
		ipInfo.IntfRef = obj.IntfRef
		ipInfo.IfIndex = obj.IfIndex
		ipInfo.OperState = obj.OperState
		v4Info.Cfg.IpAddr = obj.IpAddr
		v4Info.Vrrpkey = nil
		//		}
		svr.V4[obj.IfIndex] = ipInfo
		svr.V4IntfRefToIfIndex[obj.IntfRef] = obj.IfIndex
	}
}

func (svr *VrrpServer) getIPv6Intfs() {
	ipv6Info, err := svr.SwitchPlugin.GetAllIPv6IntfState()
	if err != nil {
		debug.Logger.Err("Failed to get all IPv6 interfaces, err:", err)
		return
	}
	for _, obj := range ipv6Info {
		// do not care for loopback interface
		if svr.SwitchPlugin.IsLoopbackType(obj.IfIndex) {
			continue
		}
		v6Info, _ := svr.V6[obj.IfIndex]
		ipInfo := v6Info.Cfg.Info
		//if !exists {
		ipInfo.IntfRef = obj.IntfRef
		ipInfo.IfIndex = obj.IfIndex
		ipInfo.OperState = obj.OperState
		ip, _, _ := net.ParseCIDR(obj.IpAddr)
		if ip.IsLinkLocalUnicast() {
			v6Info.Cfg.LinkScopeAddr = ip.String()
		} else {
			v6Info.Cfg.IPv6Addr = ip.String()
		}
		//		}
		v6Info.Vrrpkey = nil
		svr.V6[obj.IfIndex] = ipInfo
		svr.V6IntfRefToIfIndex[obj.IntfRef] = obj.IfIndex
	}
}

func (svr *VrrpServer) GetIPIntfs() {

	debug.Logger.Info("Get all ipv4 interfaces from asicd")
	svr.getIPv4Intfs()
	debug.Logger.Info("Get all ipv6 interfaces from asicd")
	svr.getIPv6Intfs()
}
