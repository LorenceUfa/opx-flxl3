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
	"errors"
	"fmt"
	"l3/vrrp/config"
	"l3/vrrp/debug"
	"l3/vrrp/packet"
	"syscall"
	"utils/netUtils"
)

func (svr *VrrpServer) GetPorts() {
	/*
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
		}

		debug.Logger.Info("Done with Port State list")
	*/
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
		v4Obj := &V4Intf{}
		ipInfo := &config.BaseIpInfo{
			IntfRef:   obj.IntfRef,
			IfIndex:   obj.IfIndex,
			OperState: obj.OperState,
			IpAddr:    obj.IpAddr,
		}
		v4Obj.Init(ipInfo)
		svr.V4[obj.IfIndex] = v4Obj
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
		v6Obj := &V6Intf{}
		ipInfo := &config.BaseIpInfo{
			IntfRef:   obj.IntfRef,
			IfIndex:   obj.IfIndex,
			OperState: obj.OperState,
			IpAddr:    obj.IpAddr,
		}
		v6Obj.Init(ipInfo)
		svr.V6[obj.IfIndex] = v6Obj
		svr.V6IntfRefToIfIndex[obj.IntfRef] = obj.IfIndex
	}
}

func (svr *VrrpServer) GetIPIntfs() {

	debug.Logger.Info("Get all ipv4 interfaces from asicd")
	svr.getIPv4Intfs()
	debug.Logger.Info("Get all ipv6 interfaces from asicd")
	svr.getIPv6Intfs()
}

func constructIntfKey(intfRef string, vrid int32, version uint8) KeyInfo {
	return KeyInfo{intfRef, vrid, version}
}

func (svr *VrrpServer) ValidateCreateConfig(cfg *config.IntfCfg) (bool, error) {
	key := constructIntfKey(cfg.IntfRef, cfg.VRID, cfg.Version)
	if _, exists := svr.Intf[key]; exists {
		return false, errors.New(fmt.Sprintln("Vrrp Interface already created for config:", cfg,
			"only update is allowed"))
	}

	// check if ipv4 address is configured on the intfRef
	_, v4exists := svr.V4IntfRefToIfIndex[cfg.IntfRef]

	// check if ipv6 address is configured on the intRef
	_, v6exists := svr.V6IntfRefToIfIndex[cfg.IntfRef]

	if !v4exists && !v6exists {
		return false, errors.New(fmt.Sprintln("Vrrp cannot be configured as no l3 Interface found for:",
			cfg.IntfRef))
	}
	debug.Logger.Info("Validation of create config:", *cfg, "is success")
	return true, nil
}

func (svr *VrrpServer) ValidateUpdateConfig(cfg *config.IntfCfg) (bool, error) {
	key := constructIntfKey(cfg.IntfRef, cfg.VRID, cfg.Version)
	intf, exists := svr.Intf[key]
	if !exists {
		return false, errors.New(fmt.Sprintln("Vrrp Interface doesn't exists for config:", cfg,
			"please do create before updating entry"))
	}
	if intf.Config.VRID != cfg.VRID {
		return false, errors.New("Updating VRID is not allowed")
	}
	return true, nil
}

func (svr *VrrpServer) ValidateDeleteConfig(cfg *config.IntfCfg) (bool, error) {
	key := constructIntfKey(cfg.IntfRef, cfg.VRID, cfg.Version)
	if _, exists := svr.Intf[key]; !exists {
		return false, errors.New(fmt.Sprintln("Vrrp Interface was not created for config:", cfg))
	}
	return true, nil
}

func (svr *VrrpServer) ValidConfiguration(cfg *config.IntfCfg) (bool, error) {
	if cfg.VRID == 0 {
		return false, errors.New(fmt.Sprintln(VRRP_INVALID_VRID, cfg.VRID))
	}
	switch cfg.Operation {
	case config.CREATE:
		return svr.ValidateCreateConfig(cfg)
	case config.UPDATE:
		return svr.ValidateUpdateConfig(cfg)
	case config.DELETE:
		return svr.ValidateDeleteConfig(cfg)
	}
	return false, errors.New("Invalid Operation received for Vrrp Interface Config")
}

/* Update Intf List which can be used during state
 */
func (svr *VrrpServer) updateIntfList(key KeyInfo, version uint8, insert bool) {
	switch insert {
	case true:
		// new vrrp configured insert the entry into lists
		if version == config.VERSION2 {
			svr.v4Intfs = append(svr.v4Intfs, key)
		} else {
			svr.v6Intfs = append(svr.v6Intfs, key)
		}
	case false:
		if version == config.VERSION2 {
			for idx, _ := range svr.v4Intfs {
				if svr.v4Intfs[idx] == key {
					svr.v4Intfs = append(svr.v4Intfs[:idx], svr.v4Intfs[idx+1:]...)
					break
				}
			}
		} else {
			for idx, _ := range svr.v6Intfs {
				if svr.v6Intfs[idx] == key {
					svr.v6Intfs = append(svr.v6Intfs[:idx], svr.v6Intfs[idx+1:]...)
					break
				}
			}
		}
	}
}

/* During Create of Virtual Interface Enable should always be set to false... when
 * vrrp interface becomes master it will request for the interface to be in up state
 * Input: (vrrp interface config, virtual mac)
 */
func (svr *VrrpServer) CreateVirtualIntf(cfg *config.IntfCfg, vMac string) {
	if svr.GlobalConfig.Enable == false {
		return
	}
	debug.Logger.Info("Vrrp Creating Virtual Interface for:", cfg.IntfRef, cfg.VirtualIPAddr, vMac)
	switch cfg.Version {
	case config.VERSION2:
		svr.SwitchPlugin.CreateVirtualIPv4Intf(cfg.IntfRef, cfg.VirtualIPAddr, vMac, false /*enable*/)
	case config.VERSION3:
		svr.SwitchPlugin.CreateVirtualIPv6Intf(cfg.IntfRef, cfg.VirtualIPAddr, vMac, false /*enable*/)
	}
}

/* Update Virtual Interface by changing the state as requested
 * Input: (intfRef, virtual ip address, macAddress, enable)
 */
func (svr *VrrpServer) UpdateVirtualIntf(virtualIpInfo *config.VirtualIpInfo) {
	switch virtualIpInfo.Version {
	case config.VERSION2:
		svr.SwitchPlugin.UpdateVirtualIPv4Intf(virtualIpInfo.IntfRef, virtualIpInfo.IpAddr, virtualIpInfo.MacAddr, virtualIpInfo.Enable)
	case config.VERSION3:
		svr.SwitchPlugin.UpdateVirtualIPv6Intf(virtualIpInfo.IntfRef, virtualIpInfo.IpAddr, virtualIpInfo.MacAddr, virtualIpInfo.Enable)
	}
}

/*
 *  Handling Vrrp Interface Configuration
 */
func (svr *VrrpServer) HandlerVrrpIntfCreateConfig(cfg *config.IntfCfg) {
	debug.Logger.Info("Received vrrp interface create config:", *cfg)
	key := constructIntfKey(cfg.IntfRef, cfg.VRID, cfg.Version)
	intf, exists := svr.Intf[key]
	if exists {
		debug.Logger.Err("During Create we should not have any entry in the DB")
		return
	}
	debug.Logger.Debug("Constructed Key for vrrp interface is:", key)
	l3Info := &config.BaseIpInfo{}
	l3Info.IntfRef = cfg.IntfRef
	var ipIntf IPIntf
	var ifIndex int32
	// Get DB based on config version
	switch cfg.Version {
	case config.VERSION2:
		ifIndex, exists = svr.V4IntfRefToIfIndex[cfg.IntfRef]
		debug.Logger.Debug("v4 ifIndex found in reverse map for:", cfg.IntfRef, "is:", ifIndex, "exists:", exists)
		if exists {
			ipIntf, exists = svr.V4[ifIndex]
		}
	// if cross reference exists then only set l3Info else just pass go defaults and it will updated
	// later once we have configured ipv4 or ipv6 interface
	case config.VERSION3:
		ifIndex, exists = svr.V6IntfRefToIfIndex[cfg.IntfRef]
		debug.Logger.Debug("v6 ifIndex found in reverse map for:", cfg.IntfRef, "is:", ifIndex, "exists:", exists)
		if exists {
			ipIntf, exists = svr.V6[ifIndex]
		}
	}
	// if entry exists then only you should get information from DB otherwise it should be nothing
	// Information collected from DB is L3 interface ip address and operation state
	if exists {
		debug.Logger.Debug("ip interface exists and hence get information from DB")
		l3Info.IfIndex = ifIndex
		ipIntf.GetObjFromDb(l3Info)
	}
	intf.InitVrrpIntf(cfg, l3Info, svr.VirtualIpCh)
	// if l3 interface was created before vrrp interface then there might be a chance that interface is already
	// up... if that's the case then lets start fsm right away
	if l3Info.OperState == config.STATE_UP && svr.GlobalConfig.Enable && cfg.AdminState {
		// during create always call start fsm
		intf.StartFsm()
	}
	svr.Intf[key] = intf
	ipIntf.SetVrrpIntfKey(key)
	debug.Logger.Info("Fsm is initialized for the interface, now calling create virtual interface")
	svr.CreateVirtualIntf(cfg, intf.GetVMac())
	svr.updateIntfList(key, cfg.Version, true /*insert*/)
}

func (svr *VrrpServer) HandleVrrpIntfUpdateConfig(cfg *config.IntfCfg) {
	debug.Logger.Info("Received interface update config:", *cfg)
	key := constructIntfKey(cfg.IntfRef, cfg.VRID, cfg.Version)
	intf, exists := svr.Intf[key]
	if !exists {
		debug.Logger.Err("Cannot perform update as no interface found in db for key:", key)
		return
	}
	intf.UpdateConfig(cfg)
	svr.Intf[key] = intf
}

func (svr *VrrpServer) HandleVrrpIntfDeleteConfig(cfg *config.IntfCfg) {
	debug.Logger.Info("Received vrrp interface create config:", *cfg)
	key := constructIntfKey(cfg.IntfRef, cfg.VRID, cfg.Version)
	intf, exists := svr.Intf[key]
	if !exists {
		// this should never happen as Validate should have taken care of this
		debug.Logger.Err("no vrrp interface found for:", key)
		return
	}
	intf.DeInitVrrpIntf()
	delete(svr.Intf, key)
	svr.updateIntfList(key, cfg.Version, false /*delete*/)
}

func (svr *VrrpServer) HandleVrrpIntfConfig(cfg *config.IntfCfg) {
	debug.Logger.Info("svr handling vrrp interface configuration:", *cfg)
	switch cfg.Operation {
	case config.CREATE:
		svr.HandlerVrrpIntfCreateConfig(cfg)
	case config.UPDATE:
		svr.HandleVrrpIntfUpdateConfig(cfg)
	case config.DELETE:
		// @TODO: need to handle delete vrrp intf config
		svr.HandleVrrpIntfDeleteConfig(cfg)
	}
}

func (svr *VrrpServer) HandleProtocolMacEntry(add bool) {
	switch add {
	case true:
		svr.SwitchPlugin.EnablePacketReception(packet.VRRP_PROTOCOL_MAC, -1, 1)
	case false:
		svr.SwitchPlugin.DisablePacketReception(packet.VRRP_PROTOCOL_MAC, -1, 1)
	}
}

/*
 * We can get vrrp interface configurations before even vrrp is enabled....let's handle that scenario here
 * by starting fsm if vrrp enabled
 * by stopping fsm if vrrp disabled
 */
func (svr *VrrpServer) HandleVrrpEnableDisable(enable bool) {
	debug.Logger.Info("vrrp globally:", enable, "handling it")
	for key, intf := range svr.Intf {
		if enable {
			intf.UpdateIpState()
		} else {
			intf.StopFsm()
		}
		svr.Intf[key] = intf
	}
}

func (svr *VrrpServer) HandleGlobalConfig(gCfg *config.GlobalConfig) {
	debug.Logger.Info("Handling Global Config for:", *gCfg)
	svr.GlobalConfig.Enable = gCfg.Enable
	switch gCfg.Operation {
	case config.CREATE:
		debug.Logger.Info("Vrrp Global Object Created")
		svr.GlobalConfig.Vrf = gCfg.Vrf
	case config.UPDATE:
		debug.Logger.Info("Vrrp Global Updated:", *svr.GlobalConfig)
		if gCfg.Enable {
			debug.Logger.Info("Vrrp Enabled, configuring Protocol Mac")
			svr.HandleProtocolMacEntry(true /*Enable*/)
		} else {
			debug.Logger.Info("Vrrp Disabled, deleting Protocol Mac")
			svr.HandleProtocolMacEntry(false /*Enable*/)
		}
		svr.HandleVrrpEnableDisable(gCfg.Enable)
	}
}

func (svr *VrrpServer) HandleIpStateChange(msg *config.BaseIpInfo) {
	var ipIntf IPIntf
	var exists bool

	switch msg.IpType {
	case syscall.AF_INET:
		ipIntf, exists = svr.V4[msg.IfIndex]
	case syscall.AF_INET6:
		ipIntf, exists = svr.V6[msg.IfIndex]
	}

	if !exists {
		debug.Logger.Err("No Entry found for:", *msg, "during state:", msg.OperState, "notification")
		return
	}
	// update sw state for ip interface with new information
	ipIntf.Update(msg)

	// check if vrrp is enabled or not
	if svr.GlobalConfig.Enable == false {
		debug.Logger.Info("Vrrp is not enabled and hence just updating ip information")
		return
	}
	// get the vrrp interface key
	key := ipIntf.GetVrrpIntfKey()
	if key == nil {
		// if no key then it means that no vrrp interface is created
		debug.Logger.Warning("No vrrp interface attached to ip interface:", ipIntf.GetIntfRef())
		return
	}
	intf, exists := svr.Intf[*key]
	if !exists {
		debug.Logger.Warning("No Vrrp Interface configured and hence nothing to do")
		return
	}
	intf.UpdateOperState(msg.OperState)
	intf.UpdateIpState()
	svr.Intf[*key] = intf
}

func (svr *VrrpServer) HandleIpNotification(msg *config.BaseIpInfo) {
	debug.Logger.Info("handling ip notification:", *msg)
	switch msg.MsgType {
	case config.IP_MSG_CREATE:
		switch msg.IpType {
		case syscall.AF_INET:
			v4, exists := svr.V4[msg.IfIndex]
			if !exists {
				v4 = &V4Intf{}
				v4.Init(msg)
				svr.V4[msg.IfIndex] = v4
				svr.V4IntfRefToIfIndex[msg.IntfRef] = msg.IfIndex
				debug.Logger.Info("Reverse v4 ip intf to ifIndex cached for:", msg.IntfRef, "-------->", msg.IfIndex)
			}
		case syscall.AF_INET6:
			v6, exists := svr.V6[msg.IfIndex]
			if !exists && !netUtils.IsIpv6LinkLocal(msg.IpAddr) {
				v6 = &V6Intf{}
				v6.Init(msg)
				svr.V6[msg.IfIndex] = v6
				svr.V6IntfRefToIfIndex[msg.IntfRef] = msg.IfIndex
				debug.Logger.Info("Reverse v6 ip intf to ifIndex cached for:", msg.IntfRef, "-------->", msg.IfIndex)
			}
		}
	case config.IP_MSG_DELETE:
		// @TODO: need to stop fsm
		switch msg.IpType {
		case syscall.AF_INET:
			v4, exists := svr.V4[msg.IfIndex]
			if exists {
				v4.DeInit(msg)
			}
		case syscall.AF_INET6:
			// most likely we will get two delete one for linkscope and other for global-scope
			v6, exists := svr.V6[msg.IfIndex]
			if exists && !netUtils.IsIpv6LinkLocal(msg.IpAddr) {
				v6.DeInit(msg)
			}
		}

	case config.IP_MSG_STATE_CHANGE:
		svr.HandleIpStateChange(msg)
	}
}
