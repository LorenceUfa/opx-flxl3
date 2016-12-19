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
	"arpdInt"
	"encoding/json"
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
	"utils/clntUtils/clntDefs/arpdClntDefs"
	"utils/ipcutils"
	"utils/logging"
)

type ArpdClient struct {
	ClientBase
	ClientHdl *arpd.ARPDServicesClient
}

type ClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

type ClientBase struct {
	Address            string
	Transport          thrift.TTransport
	PtrProtocolFactory *thrift.TBinaryProtocolFactory
}

var arpdMutex *sync.Mutex = &sync.Mutex{}
var Logger logging.LoggerIntf

func GetArpdThriftClientHdl(paramsFile string, logger logging.LoggerIntf) *arpd.ARPDServicesClient {
	var arpdClient ArpdClient
	Logger = logger
	logger.Debug("Inside connectToServers...paramsFile", paramsFile)
	var clientsList []ClientJson

	bytes, err := ioutil.ReadFile(paramsFile)
	if err != nil {
		logger.Err("Error in reading configuration file")
		return nil
	}

	err = json.Unmarshal(bytes, &clientsList)
	if err != nil {
		logger.Err("Error in Unmarshalling Json")
		return nil
	}

	for _, client := range clientsList {
		if client.Name == "arpd" {
			logger.Debug("found arpd at port", client.Port)
			arpdClient.Address = "localhost:" + strconv.Itoa(client.Port)
			arpdClient.Transport, arpdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(arpdClient.Address)
			if err != nil {
				logger.Err("Failed to connect to Arpd, retrying until connection is successful")
				count := 0
				ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
				for _ = range ticker.C {
					arpdClient.Transport, arpdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(arpdClient.Address)
					if err == nil {
						ticker.Stop()
						break
					}
					count++
					if (count % 10) == 0 {
						logger.Err("Still can't connect to Arpd, retrying..")
					}
				}

			}
			logger.Info("Connected to Arpd")
			arpdClient.ClientHdl = arpd.NewARPDServicesClientFactory(arpdClient.Transport, arpdClient.PtrProtocolFactory)
			return arpdClient.ClientHdl
		}
	}
	return nil
}

func (arpdClientMgr *FSArpdClntMgr) GetArpEntryState(ipAddr string) (*arpdClntDefs.ArpEntryState, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		obj, err := arpdClientMgr.ClientHdl.GetArpEntryState(ipAddr)
		arpdMutex.Unlock()
		if err != nil {
			return nil, err
		}
		retObj := convertFromThriftArpEntryState(obj)
		return retObj, nil
	}
	return nil, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) GetBulkArpEntryState(fromIdx, count int) (*arpdClntDefs.ArpEntryStateGetInfo, error) {
	if arpdClientMgr.ClientHdl == nil {
		return nil, errors.New("Arpd Client Handle is nil")
	}
	arpdMutex.Lock()
	bulkInfo, err := arpdClientMgr.ClientHdl.GetBulkArpEntryState(arpd.Int(fromIdx), arpd.Int(count))
	arpdMutex.Unlock()
	if bulkInfo == nil {
		return nil, err
	}
	var retObj arpdClntDefs.ArpEntryStateGetInfo
	retObj.StartIdx = int32(bulkInfo.StartIdx)
	retObj.EndIdx = int32(bulkInfo.EndIdx)
	retObj.Count = int32(bulkInfo.Count)
	retObj.More = bulkInfo.More
	retObj.ArpEntryStateList = make([]*arpdClntDefs.ArpEntryState, int(retObj.Count))
	for idx := 0; idx < int(retObj.Count); idx++ {
		retObj.ArpEntryStateList[idx] = convertFromThriftArpEntryState(bulkInfo.ArpEntryStateList[idx])
	}
	return &retObj, nil
}

func (arpdClientMgr *FSArpdClntMgr) GetArpLinuxEntryState(ipAddr string) (*arpdClntDefs.ArpLinuxEntryState, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		obj, err := arpdClientMgr.ClientHdl.GetArpLinuxEntryState(ipAddr)
		arpdMutex.Unlock()
		if err != nil {
			return nil, err
		}
		retObj := convertFromThriftArpLinuxEntryState(obj)
		return retObj, nil
	}
	return nil, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) GetBulkArpLinuxEntryState(fromIdx, count int) (*arpdClntDefs.ArpLinuxEntryStateGetInfo, error) {
	if arpdClientMgr.ClientHdl == nil {
		return nil, errors.New("Arpd Client Handle is nil")
	}
	arpdMutex.Lock()
	bulkInfo, err := arpdClientMgr.ClientHdl.GetBulkArpLinuxEntryState(arpd.Int(fromIdx), arpd.Int(count))
	arpdMutex.Unlock()
	if bulkInfo == nil {
		return nil, err
	}
	var retObj arpdClntDefs.ArpLinuxEntryStateGetInfo
	retObj.StartIdx = int32(bulkInfo.StartIdx)
	retObj.EndIdx = int32(bulkInfo.EndIdx)
	retObj.Count = int32(bulkInfo.Count)
	retObj.More = bulkInfo.More
	retObj.ArpLinuxEntryStateList = make([]*arpdClntDefs.ArpLinuxEntryState, int(retObj.Count))
	for idx := 0; idx < int(retObj.Count); idx++ {
		retObj.ArpLinuxEntryStateList[idx] = convertFromThriftArpLinuxEntryState(bulkInfo.ArpLinuxEntryStateList[idx])
	}
	return &retObj, nil
}

func (arpdClientMgr *FSArpdClntMgr) ExecuteActionArpDeleteByIfName(cfg *arpdClntDefs.ArpDeleteByIfName) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convCfg := convertToThriftArpDeleteByIfName(cfg)
		retVal, err := arpdClientMgr.ClientHdl.ExecuteActionArpDeleteByIfName(convCfg)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) ExecuteActionArpDeleteByIPv4Addr(cfg *arpdClntDefs.ArpDeleteByIPv4Addr) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convCfg := convertToThriftArpDeleteByIPv4Addr(cfg)
		retVal, err := arpdClientMgr.ClientHdl.ExecuteActionArpDeleteByIPv4Addr(convCfg)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) ExecuteActionArpRefreshByIfName(cfg *arpdClntDefs.ArpRefreshByIfName) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convCfg := convertToThriftArpRefreshByIfName(cfg)
		retVal, err := arpdClientMgr.ClientHdl.ExecuteActionArpRefreshByIfName(convCfg)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) ExecuteActionArpRefreshByIPv4Addr(cfg *arpdClntDefs.ArpRefreshByIPv4Addr) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convCfg := convertToThriftArpRefreshByIPv4Addr(cfg)
		retVal, err := arpdClientMgr.ClientHdl.ExecuteActionArpRefreshByIPv4Addr(convCfg)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) CreateArpGlobal(cfg *arpdClntDefs.ArpGlobal) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convCfg := convertToThriftArpGlobal(cfg)
		retVal, err := arpdClientMgr.ClientHdl.CreateArpGlobal(convCfg)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) UpdateArpGlobal(origCfg, newCfg *arpdClntDefs.ArpGlobal, attrset []bool, op []*arpdClntDefs.PatchOpInfo) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convOrigCfg := convertToThriftArpGlobal(origCfg)
		convNewCfg := convertToThriftArpGlobal(newCfg)
		convOp := convertToThriftPatchOpInfo(op)
		retVal, err := arpdClientMgr.ClientHdl.UpdateArpGlobal(convOrigCfg, convNewCfg, attrset, convOp)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) DeleteArpGlobal(cfg *arpdClntDefs.ArpGlobal) (bool, error) {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		convCfg := convertToThriftArpGlobal(cfg)
		retVal, err := arpdClientMgr.ClientHdl.DeleteArpGlobal(convCfg)
		arpdMutex.Unlock()
		return retVal, err
	}
	return false, errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) ResolveArpIPv4(destNetIp string, ifIdx int32) error {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		err := arpdClientMgr.ClientHdl.ResolveArpIPv4(destNetIp, arpdInt.Int(ifIdx))
		arpdMutex.Unlock()
		return err
	}
	return errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) DeleteResolveArpIPv4(NbrIP string) error {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		err := arpdClientMgr.ClientHdl.DeleteResolveArpIPv4(NbrIP)
		arpdMutex.Unlock()
		return err
	}
	return errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) DeleteArpEntry(ipAddr string) error {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		err := arpdClientMgr.ClientHdl.DeleteArpEntry(ipAddr)
		arpdMutex.Unlock()
		return err
	}
	return errors.New("Arpd Client Handle is nil")
}

func (arpdClientMgr *FSArpdClntMgr) SendGarp(ifName string, macAddr string, ipAddr string) error {
	if arpdClientMgr.ClientHdl != nil {
		arpdMutex.Lock()
		err := arpdClientMgr.ClientHdl.SendGarp(ifName, macAddr, ipAddr)
		arpdMutex.Unlock()
		return err
	}
	return errors.New("Arpd Client Handle is nil")
}
