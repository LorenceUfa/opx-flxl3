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
	"l3/arp/clientMgr"
	arpFS "l3/arp/clientMgr/flexswitch"
	ndpClient "l3/ndp/lib"
	ndpFS "l3/ndp/lib/flexswitch"
	"l3/vrrp/debug"
	"log/syslog"
	"testing"
	asicdClnt "utils/asicdClient/mock"
	"utils/commonDefs"
	"utils/logging"
)

var testSvr *VrrpServer

func TestServerInit(t *testing.T) {
	var err error
	logger := new(logging.Writer)
	logger.MyComponentName = "VRRPD"
	logger.SysLogger, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "VRRPFSMTEST")
	if err != nil {
		t.Error("failed to initialize syslog:", err)
		return
	} else {
		logger.MyLogLevel = 9 // trace level
		debug.SetLogger(logger)
	}
	testAsicdClnt := &asicdClnt.MockAsicdClientMgr{}
	testNdpInst := ndpFS.NdpdClientStruct{}
	testArpInst := arpFS.ArpdClientStruct{}
	testArpClnt := arpClient.NewArpdClient(commonDefs.MOCK_PLUGIN, "", make([]commonDefs.ClientJson, 0), testArpInst)
	testNdpClnt := ndpClient.NewNdpClient(commonDefs.MOCK_PLUGIN, "", make([]commonDefs.ClientJson, 0), testNdpInst)
	testSvr = VrrpNewServer(testAsicdClnt, testArpClnt, testNdpClnt, nil)
	if testSvr == nil {
		t.Error("Initializing Vrrp Server Failed")
		return
	}
}
