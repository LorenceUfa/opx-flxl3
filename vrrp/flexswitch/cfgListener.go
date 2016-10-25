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

package flexswitch

import (
	"errors"
	"l3/vrrp/api"
	"l3/vrrp/config"
	"l3/vrrp/debug"
	"vrrpd"
)

func (h *ConfigHandler) CreateVrrpGlobal(config *vrrpd.VrrpGlobal) (r bool, err error) {
	debug.Logger.Debug("Thrift request for creating vrrp global object:", *gblCfg)
	gblCfg := &config.GlobalConfig{config.Vrf, config.Enable, config.CREATE}
	api.CreateVrrpGbl(gblCfg)
	debug.Logger.Debug("Thrift returning for creating vrrp global object true, nil")
	return true, nil
}

func (h *ConfigHandler) UpdateVrrpGlobal(config *vrrpd.VrrpGlobal) (r bool, err error) {
	debug.Logger.Debug("Thrift request for creating vrrp global object:", *gblCfg)
	gblCfg := &config.GlobalConfig{config.Vrf, config.Enable, config.UPDATE}
	api.UpdateVrrpGbl(gblCfg)
	debug.Logger.Debug("Thrift returning for creating vrrp global object true, nil")
	return true, nil
}

func (h *ConfigHandler) DeleteVrrpGlobal(config *vrrpd.VrrpGlobal) (r bool, err error) {
	debug.Logger.Debug("Thrift request for deleting vrrp global object:", *gblCfg)
	err = errors.New("Deleting Vrrp Global Object is not Supported")
	r = false
	debug.Logger.Debug("Thrift returning for deleting vrrp global object:", r, err)
	return r, err
}

func (h *ConfigHandler) CreateVrrpV4Intf(config *vrrpd.VrrpV4Intf) (r bool, err error) {
	debug.Logger.Debug("Thrift request received for creating vrrp v4 interface config for:", *config)
	v4Cfg := &config.IntfCfg{
		IntfRef:               config.IntfRef,
		VRID:                  config.VRID,
		Priority:              config.Priority,
		VirtualIPv4Addr:       config.VirtualIPv4Addr,
		AdvertisementInterval: config.AdvertisementInterval,
		PreemptMode:           config.PreemptMode,
		AcceptMode:            config.AcceptMode,
		Version:               config.VERSION2,
		Operation:             config.CREATE,
	}
	r, err = api.VrrpIntfConfig(v4Cfg)
	debug.Logger.Debug("Thrift request returning for creating vrrp v4 interface config returning:", r, err)
	return r, err
}
func (h *ConfigHandler) UpdateVrrpV4Intf(origconfig *vrrpd.VrrpV4Intf, newconfig *vrrpd.VrrpV4Intf, attrset []bool, op []*vrrpd.PatchOpInfo) (r bool, err error) {
	debug.Logger.Debug("Thrift request received for updating vrrp v4 interface config for:", origconfig, "to new:", newconfig)
	v4Cfg := &config.IntfCfg{
		IntfRef:               config.IntfRef,
		VRID:                  config.VRID,
		Priority:              config.Priority,
		VirtualIPv4Addr:       config.VirtualIPv4Addr,
		AdvertisementInterval: config.AdvertisementInterval,
		PreemptMode:           config.PreemptMode,
		AcceptMode:            config.AcceptMode,
		Version:               config.VERSION2,
		Operation:             config.UPDATE,
	}
	r, err = api.VrrpIntfConfig(v4Cfg)
	debug.Logger.Debug("Thrift request returning for updating vrrp v4 interface config for:", origconfig, "to new:",
		newconfig, "returning:", r, err)
	return true, nil
}

func (h *ConfigHandler) DeleteVrrpV4Intf(config *vrrpd.VrrpV4Intf) (r bool, err error) {
	debug.Logger.Debug("Thrift request received for deleting vrrp v4 interface config for:", *config)
	v4Cfg := &config.IntfCfg{
		IntfRef:               config.IntfRef,
		VRID:                  config.VRID,
		Priority:              config.Priority,
		VirtualIPv4Addr:       config.VirtualIPv4Addr,
		AdvertisementInterval: config.AdvertisementInterval,
		PreemptMode:           config.PreemptMode,
		AcceptMode:            config.AcceptMode,
		Version:               config.VERSION2,
		Operation:             config.DELETE,
	}
	r, err = api.VrrpIntfConfig(v4Cfg)
	debug.Logger.Debug("Thrift request returning for deleting vrrp v4 interface config returning:", r, err)
	return r, err
}

func (h *ConfigHandler) CreateVrrpV6Intf(config *vrrpd.VrrpV6Intf) (r bool, err error) {
	debug.Logger.Debug("Thrift request received for creating vrrp v6 interface config for:", *config)
	v6Cfg := &config.IntfCfg{
		IntfRef:               config.IntfRef,
		VRID:                  config.VRID,
		Priority:              config.Priority,
		VirtualIPv4Addr:       config.VirtualIPv4Addr,
		AdvertisementInterval: config.AdvertisementInterval,
		PreemptMode:           config.PreemptMode,
		AcceptMode:            config.AcceptMode,
		Version:               config.VERSION3,
		Operation:             config.CREATE,
	}
	r, err = api.VrrpIntfConfig(v6Cfg)
	debug.Logger.Debug("Thrift request returning for creating vrrp v6 interface config returning:", r, err)
	return r, err
}

func (h *ConfigHandler) UpdateVrrpV6Intf(origconfig *vrrpd.VrrpV6Intf, newconfig *vrrpd.VrrpV6Intf, attrset []bool, op []*vrrpd.PatchOpInfo) (r bool, err error) {
	debug.Logger.Debug("Thrift request received for updating vrrp v6 interface config for:", origconfig, "to new:", newconfig)
	v6Cfg := &config.IntfCfg{
		IntfRef:               config.IntfRef,
		VRID:                  config.VRID,
		Priority:              config.Priority,
		VirtualIPv4Addr:       config.VirtualIPv4Addr,
		AdvertisementInterval: config.AdvertisementInterval,
		PreemptMode:           config.PreemptMode,
		AcceptMode:            config.AcceptMode,
		Version:               config.VERSION3,
		Operation:             config.UPDATE,
	}
	r, err = api.VrrpIntfConfig(v6Cfg)
	debug.Logger.Debug("Thrift request returning for updating vrrp v6 interface config for:", origconfig, "to new:",
		newconfig, "returning:", r, err)
	return true, nil
}

func (h *ConfigHandler) DeleteVrrpV6Intf(config *vrrpd.VrrpV6Intf) (r bool, err error) {
	debug.Logger.Debug("Thrift request received for deleting vrrp v6 interface config for:", *config)
	v6Cfg := &config.IntfCfg{
		IntfRef:               config.IntfRef,
		VRID:                  config.VRID,
		Priority:              config.Priority,
		VirtualIPv4Addr:       config.VirtualIPv4Addr,
		AdvertisementInterval: config.AdvertisementInterval,
		PreemptMode:           config.PreemptMode,
		AcceptMode:            config.AcceptMode,
		Version:               config.VERSION3,
		Operation:             config.DELETE,
	}
	r, err = api.VrrpIntfConfig(v6Cfg)
	debug.Logger.Debug("Thrift request returning for deleting vrrp v6 interface config returning:", r, err)
	return r, err
}
