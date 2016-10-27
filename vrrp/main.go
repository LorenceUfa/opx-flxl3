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
		status := vrrpBase.Init()
		if status == false {
			fmt.Println("Failed to do daemon base init for VRRP")
			return
		}
		// create handler and map for receiving notifications from asicd
		asicdHdl := flexswitch.GetSwitchInst()
		asicdHdl.Logger = vrrpBase.GetLogger()
		debug.SetLogger(vrrpBase.GetLogger())
		debug.Logger.Info("Initializing switch plugin")
		// connect to server and do the initializing
		switchPlugin := vrrpBase.InitSwitch("Flexswitch", "vrrpd", "VRRP", *asicdHdl)
		// create north bound config listener for clients
		debug.Logger.Info("Creating Config Plugin")
		cfgPlugin := flexswitch.NewConfigPlugin(flexswitch.NewConfigHandler(), vrrpBase.ParamsDir)
		// create new vrrp server and cache the information for switch/asicd plugin
		vrrpSvr := server.VrrpNewServer(switchPlugin, vrrpBase)
		// Init API layer after server is created
		api.Init(vrrpSvr)
		// build basic VRRP Server Information
		debug.Logger.Info("Starting VRRP Server")
		vrrpSvr.VrrpStartServer()
		vrrpBase.StartKeepAlive()
		debug.Logger.Info("Starting Config Listener for FlexSwitch Plugin")
		cfgPlugin.StartConfigListener()
		/*
			paramsDir := flag.String("params", "./params", "Params directory")
			flag.Parse()
			fileName := *paramsDir
			if fileName[len(fileName)-1] != '/' {
				fileName = fileName + "/"
			}

			fmt.Println("Start logger")
			logger, err := logging.NewLogger("vrrpd", "VRRP", true)
			if err != nil {
				fmt.Println("Failed to start the logger. Nothing will be logged...")
			}
			logger.Info("Started the logger successfully.")

			logger.Info("Starting VRRP server....")
			// Create vrrp server handler
			vrrpSvr := vrrpServer.VrrpNewServer(logger)
			// Until Server is connected to clients do not start with RPC
			vrrpSvr.VrrpStartServer(*paramsDir)

			// Start keepalive routine
			go keepalive.InitKeepAlive("vrrpd", fileName)

			// Create vrrp rpc handler
			vrrpHdl := vrrpRpc.VrrpNewHandler(vrrpSvr, logger)
			logger.Info("Starting VRRP RPC listener....")
			err = vrrpRpc.VrrpRpcStartServer(logger, vrrpHdl, *paramsDir)
			if err != nil {
				logger.Err(fmt.Sprintln("VRRP: Cannot start vrrp server", err))
				return
			}
		*/
	}
}
