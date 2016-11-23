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

package fsm

import (
	"l3/vrrp/common"
	"l3/vrrp/debug"
	"log/syslog"
	"reflect"
	"syscall"
	"testing"
	"time"
	"utils/logging"
)

var testFsm *FSM
var testv6IntfCfg = &common.IntfCfg{
	IntfRef:               "lo",
	IfIndex:               0,
	VRID:                  1,
	Priority:              150,
	VirtualIPAddr:         "fe80::172:16:0:1/64",
	AdvertisementInterval: 3,
	PreemptMode:           true,
	AcceptMode:            false,
	AdminState:            true,
	Version:               common.VERSION3,
	Operation:             1,
	IpType:                syscall.AF_INET6,
}
var testL3Info = &common.BaseIpInfo{33554581, "lo", "fe80::c837:abff:febe:ad4", "UP", "", syscall.AF_INET6}
var testVipCh = make(chan *common.VirtualIpInfo)
var testRxCh = make(chan struct{})
var testTxCh = make(chan struct{})
var testV6Vmac = "00-00-5E-00-02-01"
var testV4Vmac = "00-00-5E-00-01-01"

func TestV6FsmInit(t *testing.T) {
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

	testFsm = InitFsm(testv6IntfCfg, testL3Info, testVipCh, testRxCh, testTxCh)
	if testFsm == nil {
		t.Error("Initializing FSM Failed")
		return
	}
}

func testFsmDeInit(t *testing.T) {
	testFsm.DeInitFsm()
	if testFsm.Config != nil {
		t.Error("failed to delete interface config")
		return
	}
	if testFsm.AdverTimer != nil {
		t.Error("failed to stop & delete advertisement timer")
		return
	}

	if testFsm.vipCh == nil {
		t.Error("return virtual ip interface channel for server shouldn't be deleted")
		return
	}

	if testFsm.stateInfo != nil {
		t.Error("Failed to delete FSM state information")
		return
	}

	if testFsm.pktCh != nil {
		t.Error("failed to delete packet channel")
		return
	}

	if testFsm.pHandle != nil {
		t.Error("failed to delte pcap handler for fsm")
	}

	if testFsm.PktInfo != nil {
		t.Error("failed to delete object for packet encoding & decoding called packetInfo")
		return
	}

	if testFsm.MasterDownTimer != nil {
		t.Error("failed to stop & delete master down timer")
		return
	}

	if testFsm.IntfEventCh != nil {
		t.Error("failed to delete interface event channel handler")
		return
	}
}

func TestV6UpdateIntfConfig(t *testing.T) {
	TestV6FsmInit(t)
	if testFsm == nil {
		return
	}
	go testFsm.StartFsm()
	testv6IntfCfg.Priority = 125
	testFsm.UpdateConfig(testv6IntfCfg)
	if !reflect.DeepEqual(testv6IntfCfg, testFsm.Config) {
		t.Error("Failed to update interface config")
		t.Error("	    Wanted Config:", *testv6IntfCfg)
		t.Error("	    Received Config:", *testFsm.Config)
		return
	}
	testFsmDeInit(t)
}
func TestFsmIsRunning(t *testing.T) {
	TestV6FsmInit(t)
	if testFsm == nil {
		return
	}
	go testFsm.StartFsm()
	time.Sleep(250 * time.Millisecond)
	if testFsm.IsRunning() == false {
		t.Error("FSM started but is running flag is set to false")
		return
	}
	testFsmDeInit(t)
}

func TestGetStateInfo(t *testing.T) {
	TestV6FsmInit(t)
	if testFsm == nil {
		return
	}
	wantStateInfo := common.State{
		IntfRef:                 testv6IntfCfg.IntfRef,
		Vrid:                    testv6IntfCfg.VRID,
		OperState:               common.STATE_DOWN,
		CurrentFsmState:         VRRP_INITIALIZE_STATE_STRING,
		MasterIp:                "",
		AdverRx:                 0,
		AdverTx:                 0,
		LastAdverRx:             "",
		LastAdverTx:             "",
		IpAddr:                  testL3Info.IpAddr,
		VirtualIp:               testv6IntfCfg.VirtualIPAddr,
		VirtualRouterMACAddress: testV6Vmac,
		AdvertisementInterval:   testv6IntfCfg.AdvertisementInterval,
		MasterDownTimer:         0,
	}
	state := common.State{}
	testFsm.GetStateInfo(&state)
	if !reflect.DeepEqual(wantStateInfo, state) {
		t.Error("Failure getting state information from fsm")
		t.Error("	    want state info:", wantStateInfo)
		t.Error("	    got state info:", state)
		return
	}
}

func TestV6VMacCreate(t *testing.T) {
	TestV6FsmInit(t)
	if testFsm == nil {
		return
	}
	testFsm.createVirtualMac()
	if testFsm.VirtualMACAddress != testV6Vmac {
		t.Error("Failed creating virtual mac for v6 interface")
		t.Error("	    wanted vmac:", testV6Vmac)
		t.Error("	    received vmac:", testFsm.VirtualMACAddress)
		return
	}
}

func TestV4VMacCreate(t *testing.T) {
	TestV6FsmInit(t)
	if testFsm == nil {
		return
	}
	testFsm.ipType = syscall.AF_INET
	testFsm.createVirtualMac()
	if testFsm.VirtualMACAddress != testV4Vmac {
		t.Error("Failed creating virtual mac for v4 interface")
		t.Error("	    wanted vmac:", testV4Vmac)
		t.Error("	    received vmac:", testFsm.VirtualMACAddress)
		return
	}
}

func TestInitPktListener(t *testing.T) {
	TestV6FsmInit(t)
	if testFsm == nil {
		return
	}
}
