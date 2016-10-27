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
	"l3/vrrp/debug"
	_ "models/objects"
	"utils/dbutils"
)

func (svr *vrrpserver) readVrrpGblCfg(dbhdl *dbutils.dbutil) {
}

func (svr *vrrpserver) readVrrpIntfCfg(dbhdl *dbutils.dbutil) {
}

func (svr *VrrpServer) ReadDB() {
	if svr.dmnBase == nil {
		return
	}
	dbHdl := svr.dmnBase.GetDbHdl()
	if dbHdl == nil {
		debug.Logger.Err("DB Handler is nil and hence cannot read anything from DATABASE")
		return
	}
	debug.Logger.Info("Reading Config from DB")
	svr.readVrrpGblCfg(dbHdl)
	svr.readVrrpIntfCfg(dbHdl)
}
