//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//       Unless required by applicable law or agreed to in writing, software
//       distributed under the License is distributed on an "AS IS" BASIS,
//       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//       See the License for the specific language governing permissions and
//       limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package fsArpdClnt

import (
	"arpd"
	"errors"
	"fmt"
	"utils/clntUtils/clntIntfs"
	"utils/logging"
)

type FSArpdClntMgr struct {
	ClientHdl *arpd.ARPDServicesClient
}

var Logger logging.LoggerIntf

func NewArpdClntInit(clntInitParams *clntIntfs.BaseClntInitParams) (*FSArpdClntMgr, error) {
	var fsArpdClntMgr FSArpdClntMgr
	var err error
	Logger = clntInitParams.Logger
	fsArpdClntMgr.ClientHdl, err = GetArpdThriftClientHdl(clntInitParams)
	if fsArpdClntMgr.ClientHdl == nil || err != nil {
		clntInitParams.Logger.Err("Unable Initialize Arpd Client", err)
		return nil, errors.New(fmt.Sprintln("Unable Initialize Arpd Client", err))
	}
	InitFSArpdSubscriber(clntInitParams)
	return &fsArpdClntMgr, nil
}
