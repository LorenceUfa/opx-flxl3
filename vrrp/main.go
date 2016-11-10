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

package main

import (
	"fmt"
	"l3/vrrp/api"
	"l3/vrrp/debug"
	"l3/vrrp/flexswitch"
	"l3/vrrp/server"
	"utils/dmnBase"
)

func main() {
	plugin := ""

	switch plugin {

	case "OvsDB":

	default:
		vrrpBase := dmnBase.NewBaseDmn("vrrpd", "VRRP")
		/*
		   err := vrrpBase.InitLogger()
		   		if err != nil {
		   			fmt.Println("Failed Initializing logger for vrrp")
		   			return
		   		}
		   		debug.SetLogger(vrrpBase.Logger)
		   		err = vrrpBase.InitDBHdl()
		   		if err != nil {
		   			debug.Logger.Warning("failed initializing db handler for vrrp no read from db")
		   		}
		*/
		status := vrrpBase.Init()
		if status == false {
			fmt.Println("Failed init basedmn for VRRP")
			return
		}
		asicdHdl := flexswitch.GetSwitchInst()
		asicdHdl.Logger = vrrpBase.Logger

		debug.Logger.Info("Initializing switch plugin")
		switchPlugin := vrrpBase.InitSwitch("Flexswitch", "vrrpd", "VRRP", *asicdHdl)

		debug.Logger.Info("Creating Config Plugin")
		cfgPlugin := flexswitch.NewConfigPlugin(flexswitch.NewConfigHandler(), vrrpBase.ParamsDir, switchPlugin)

		vrrpSvr := server.VrrpNewServer(switchPlugin, vrrpBase)

		api.Init(vrrpSvr)
		debug.Logger.Info("Starting VRRP Server")

		vrrpSvr.VrrpStartServer()

		vrrpBase.StartKeepAlive()

		debug.Logger.Info("Starting Config Listener for FlexSwitch Plugin")

		cfgPlugin.StartConfigListener()
	}
}
