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
)

func (svr *VrrpServer) ValidateCreateConfig(cfg *config.IntfCfg) (bool, error) {
	key := KeyInfo{cfg.IntfRef, cfg.VRID, cfg.Version}
	if _, exists := svr.Intf[key]; exists {
		return false, errors.New(fmt.Sprintln("Vrrp Interface already created for config:", cfg,
			"only update is allowed"))
	}
	return true, nil
}

func (svr *VrrpServer) ValidateUpdateConfig(cfg *config.IntfCfg) (bool, error) {
	key := KeyInfo{cfg.IntfRef, cfg.VRID, cfg.Version}
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
	key := KeyInfo{cfg.IntfRef, cfg.VRID, cfg.Version}
	if _, exists := svr.Intf[key]; !exists {
		return false, errors.New(fmt.Sprintln("Vrrp Interface was not created for config:", cfg))
	}
	return true, nil
}

func (svr *VrrpServer) ValidConfiguration(cfg *config.IntfCfg) (bool, error) {
	if cfg.VRID == 0 {
		return false, errors.New(fmt.Sprintln(server.VRRP_INVALID_VRID, config.VRID))
	}
	switch config.Operation {
	// @TODO: jgheewala need to handle verification during the specific operations
	case config.CREATE:
		return svr.ValidateCreateConfig(cfg)
	case config.UPDATE:
		return svr.ValidateUpdateConfig(cfg)
	case config.DELETE:
		return svr.ValidateDeleteConfig(cfg)
	}
	return false, errors.New("Invalid Operation received for Vrrp Interface Config")

}

func (svr *VrrpServer) HandlerCreateConfig(cfg *config.IntfCfg) {
	key := KeyInfo{cfg.IntfRef, cfg.VRID, cfg.Version}
	intf, exists := svr.Intf[key]
	if exists {
		debug.Logger.Err("During Create we should not any entry in the DB")
		return
	}
	l3Info := &L3Intf{}
	l3Info.IfName = cfg.IntfRef
	switch cfg.Version {
	case config.VERSION2:
		ifIndex, exists := svr.V4IntfRefToIfIndex[cfg.IntfRef]
		if exists {
			l3Info.IfIndex = ifIndex
			v4, exists := svr.V4[ifIndex]
			if exists {
				l3Info.IpAddr = v4.Cfg.IpAddr
			}
		}
		// if cross reference exists then only set l3Info else just pass go defaults and it will updated
		// later once we have configured ipv4 or ipv6 interface
	case config.VERSION3:
		ifIndex, exists := svr.V6IntfRefToIfIndex[cfg.IntfRef]
		l3Info.IfIndex = ifIndex
		if exists {
			v6, exists := svr.V6[ifIndex]
			if exists {
				// @TODO: do we have to use linkscope ip or global scope ip Check RFC
				l3Info.IpAddr = v6.Cfg.IpAddr
			}
		}
	}
	intf.InitVrrpIntf(cfg, l3Info, svr.StateCh)
	svr.Intf[key] = intf
	// @TODO: NEED TO ADD PRE - PROCESSOR SUB INTERFACE OBJECT
}

func (svr *VrrpServer) HandleIntfConfig(cfg *config.IntfCfg) {
	switch cfg.Operation {
	case config.CREATE:
		svr.HandlerCreateConfig(cfg)
	case config.UPDATE:

	case config.DELETE:
	}
}

func (svr *VrrpServer) UpdateProtocolMacEntry(add bool) {
	switch add {
	case true:
		svr.SwitchPlugin.EnablePacketReception(packet.VRRP_PROTOCOL_MAC)
	case false:
		svr.SwitchPlugin.DisablePacketReception(packet.VRRP_PROTOCOL_MAC)
		// @TODO: tear down all fsm states and configuration
	}
}

func (svr *VrrpServer) HandleGlobalConfig(gCfg *config.GlobalConfig) {
	debug.Logger.Info("Handling Global Config for:", *gCfg)
	switch gCfg.Operation {
	case config.CREATE:
		debug.Logger.Info("Vrrp Enabled, configuring Protocol Mac")
		svr.UpdateProtocolMacEntry(true /*Enable*/)
	case config.UPDATE:
		debug.Logger.Info("Vrrp Disabled, deleting Protocol Mac")
		svr.UpdateProtocolMacEntry(false /*Enable*/)
	}
}

func (svr *VrrpServer) HandleStateUpdate(intfSt *IntfState) {
	intf, exists := svr.Intf[intfSt.Key]
	if !exists {
		return
	}
	intf.UpdateStateInfo(intfSt.State)
	svr.Intf[key] = intf
}
