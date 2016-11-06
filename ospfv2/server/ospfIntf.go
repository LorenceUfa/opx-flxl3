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

package server

import (
	"errors"
	"fmt"
	"l3/ospfv2/objects"
	"net"
	"sync"
	"time"
)

type IntfConfKey struct {
	IpAddr  uint32
	IntfIdx uint32
}

type IntfConf struct {
	AdminState      bool
	AreaId          uint32
	Type            uint8
	RtrPriority     uint8
	TransitDelay    uint16
	RetransInterval uint16
	HelloInterval   uint16
	RtrDeadInterval uint32
	Cost            uint32
	Mtu             uint32

	DRIpAddr  uint32
	DRtrId    uint32
	BDRIpAddr uint32
	BDRtrId   uint32

	State               uint8
	FSMCtrlCh           chan bool
	FSMCtrlReplyCh      chan bool
	HelloIntervalTicker *time.Ticker
	WaitTimer           *time.Timer

	BackupSeenCh chan BackupSeenMsg

	NeighborMap      map[NeighborConfKey]NeighborData
	NeighCreateCh    chan NeighCreateMsg
	NeighChangeCh    chan NeighChangeMsg
	NbrStateChangeCh chan NbrStateChangeMsg
	NbrFullStateCh   chan NbrFullStateMsg

	LsaCount   uint32
	IfName     string
	IpAddr     uint32
	IfMacAddr  net.HardwareAddr
	Netmask    uint32
	txHdl      IntfTxHandle
	rxHdl      IntfRxHandle
	txHdlMutex sync.Mutex
}

func getOspfv2IntfUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0

	if attrset == nil {
		mask = objects.OSPFV2_INTF_UPDATE_ADMIN_STATE |
			objects.OSPFV2_INTF_UPDATE_AREA_ID |
			objects.OSPFV2_INTF_UPDATE_TYPE |
			objects.OSPFV2_INTF_UPDATE_RTR_PRIORITY |
			objects.OSPFV2_INTF_UPDATE_TRANSIT_DELAY |
			objects.OSPFV2_INTF_UPDATE_RETRANS_INTERVAL |
			objects.OSPFV2_INTF_UPDATE_HELLO_INTERVAL |
			objects.OSPFV2_INTF_UPDATE_RTR_DEAD_INTERVAL |
			objects.OSPFV2_INTF_UPDATE_METRIC_VALUE
	} else {
		for idx, val := range attrset {
			if true == val {
				switch idx {
				case 0:
					// IPAddress
				case 1:
					//AddressLessIfIdx
				case 2:
					mask |= objects.OSPFV2_INTF_UPDATE_ADMIN_STATE
				case 3:
					mask |= objects.OSPFV2_INTF_UPDATE_AREA_ID
				case 4:
					mask |= objects.OSPFV2_INTF_UPDATE_TYPE
				case 5:
					mask |= objects.OSPFV2_INTF_UPDATE_RTR_PRIORITY
				case 6:
					mask |= objects.OSPFV2_INTF_UPDATE_TRANSIT_DELAY
				case 7:
					mask |= objects.OSPFV2_INTF_UPDATE_RETRANS_INTERVAL
				case 8:
					mask |= objects.OSPFV2_INTF_UPDATE_HELLO_INTERVAL
				case 9:
					mask |= objects.OSPFV2_INTF_UPDATE_RTR_DEAD_INTERVAL
				case 10:
					mask |= objects.OSPFV2_INTF_UPDATE_METRIC_VALUE
				}
			}
		}
	}
	return mask
}

func (server *OSPFV2Server) updateIntf(newCfg, oldCfg *objects.Ospfv2Intf, attrset []bool) (bool, error) {
	server.logger.Info("Intf configuration update")
	intfConfKey := IntfConfKey{
		IpAddr:  newCfg.IpAddress,
		IntfIdx: newCfg.AddressLessIfIdx,
	}
	intfConfEnt, exist := server.IntfConfMap[intfConfKey]
	if !exist {
		server.logger.Err("Ospf Interface configuration doesnot exist")
		return false, errors.New("Ospf Interface configuration doesnot exist")
	}
	oldIntfConfEnt := intfConfEnt
	mask := getOspfv2IntfUpdateMask(attrset)
	if mask&objects.OSPFV2_INTF_UPDATE_ADMIN_STATE == objects.OSPFV2_INTF_UPDATE_ADMIN_STATE {
		intfConfEnt.AdminState = newCfg.AdminState
	}
	if mask&objects.OSPFV2_INTF_UPDATE_AREA_ID == objects.OSPFV2_INTF_UPDATE_AREA_ID {
		// TODO: Check if Area Exist or Not
		_, exist := server.AreaConfMap[newCfg.AreaId]
		if !exist {
			server.logger.Err("Area doesnot exist")
			return false, errors.New("Area doesnot exist")
		}
		intfConfEnt.AreaId = newCfg.AreaId
	}
	if mask&objects.OSPFV2_INTF_UPDATE_TYPE == objects.OSPFV2_INTF_UPDATE_TYPE {
		intfConfEnt.Type = newCfg.Type
	}
	if mask&objects.OSPFV2_INTF_UPDATE_RTR_PRIORITY == objects.OSPFV2_INTF_UPDATE_RTR_PRIORITY {
		intfConfEnt.RtrPriority = newCfg.RtrPriority
	}
	if mask&objects.OSPFV2_INTF_UPDATE_TRANSIT_DELAY == objects.OSPFV2_INTF_UPDATE_TRANSIT_DELAY {
		intfConfEnt.TransitDelay = newCfg.TransitDelay
	}
	if mask&objects.OSPFV2_INTF_UPDATE_RETRANS_INTERVAL == objects.OSPFV2_INTF_UPDATE_RETRANS_INTERVAL {
		intfConfEnt.RetransInterval = newCfg.RetransInterval
	}
	if mask&objects.OSPFV2_INTF_UPDATE_HELLO_INTERVAL == objects.OSPFV2_INTF_UPDATE_HELLO_INTERVAL {
		intfConfEnt.HelloInterval = newCfg.HelloInterval
	}
	if mask&objects.OSPFV2_INTF_UPDATE_RTR_DEAD_INTERVAL == objects.OSPFV2_INTF_UPDATE_RTR_DEAD_INTERVAL {
		intfConfEnt.RtrDeadInterval = newCfg.RtrDeadInterval
	}
	if mask&objects.OSPFV2_INTF_UPDATE_METRIC_VALUE == objects.OSPFV2_INTF_UPDATE_METRIC_VALUE {
		intfConfEnt.Cost = uint32(newCfg.MetricValue)
	}
	if intfConfEnt.AreaId != oldIntfConfEnt.AreaId {
		oldAreaEnt, _ := server.AreaConfMap[oldIntfConfEnt.AreaId]
		delete(oldAreaEnt.IntfMap, intfConfKey)
		server.AreaConfMap[oldIntfConfEnt.AreaId] = oldAreaEnt
		// TODO: Regenerate LSAs
	}
	if oldIntfConfEnt.AdminState == true &&
		server.globalData.AdminState == true {
		//TODO
		//Stop Interface FSM
		//Flush all the routes learned via this interface
		if oldIntfConfEnt.AreaId != intfConfEnt.AreaId {
			//Delete Interface from Old AreaId and Add Interface to New AreaId
		}
	}
	server.IntfConfMap[intfConfKey] = intfConfEnt
	if intfConfEnt.AdminState == true &&
		server.globalData.AdminState == true {
		//TODO
		//Start Interface FSM
		//Regenerate Router LSA
	} else if intfConfEnt.AdminState == false &&
		server.globalData.AdminState == true {
		//Regenerate Router LSA
	}
	return true, nil
}

func (server *OSPFV2Server) createIntf(cfg *objects.Ospfv2Intf) (bool, error) {
	var err error
	server.logger.Info("Intf configuration create")
	intfConfKey := IntfConfKey{
		IpAddr:  cfg.IpAddress,
		IntfIdx: cfg.AddressLessIfIdx,
	}

	intfConfEnt, exist := server.IntfConfMap[intfConfKey]
	if exist {
		server.logger.Err("Ospf Interface configuration already exist")
		return false, errors.New("Ospf Interface configuration already exist")
	}

	l3IfIdx, exist := server.infraData.ipToIfIdxMap[cfg.IpAddress]
	if !exist {
		// TODO: May be un numbered
		/*
			intfConfEnt.Mtu = uint32(1500) // Revisit
			intfConfEnt.IfName = ipEnt.IfName
			intfConfEnt.IfMacAddr = ipEnt.MacAddr
			intfConfEnt.Netmask = ipEnt.NetMask
		*/
	} else {
		ipEnt, _ := server.infraData.ipPropertyMap[l3IfIdx]
		if ipEnt.State == false &&
			cfg.AdminState == true &&
			server.globalData.AdminState == true {
			server.logger.Err("Ip interface is down")
			return false, errors.New("Ip Interface is down")
		}
		intfConfEnt.Mtu = uint32(ipEnt.Mtu)
		intfConfEnt.IfName = ipEnt.IfName
		intfConfEnt.IfMacAddr = ipEnt.MacAddr
		intfConfEnt.Netmask = ipEnt.NetMask
		intfConfEnt.IpAddr = ipEnt.IpAddr
	}
	intfConfEnt.AdminState = cfg.AdminState
	areaEnt, exist := server.AreaConfMap[cfg.AreaId]
	if !exist {
		server.logger.Err("Area doesnot exist")
		return false, errors.New("Area doesnot exist")
	}
	intfConfEnt.AreaId = cfg.AreaId
	intfConfEnt.Type = cfg.Type
	intfConfEnt.RtrPriority = cfg.RtrPriority
	intfConfEnt.TransitDelay = cfg.TransitDelay
	intfConfEnt.RetransInterval = cfg.RetransInterval
	intfConfEnt.HelloInterval = cfg.HelloInterval
	intfConfEnt.RtrDeadInterval = cfg.RtrDeadInterval
	intfConfEnt.Cost = uint32(cfg.MetricValue)

	intfConfEnt.DRIpAddr = 0
	intfConfEnt.DRtrId = 0
	intfConfEnt.BDRIpAddr = 0
	intfConfEnt.BDRtrId = 0

	intfConfEnt.State = objects.INTF_FSM_STATE_UNKNOWN

	intfConfEnt.FSMCtrlCh = make(chan bool)
	intfConfEnt.FSMCtrlReplyCh = make(chan bool)
	intfConfEnt.HelloIntervalTicker = nil
	intfConfEnt.WaitTimer = nil

	intfConfEnt.BackupSeenCh = make(chan BackupSeenMsg)

	intfConfEnt.NeighborMap = make(map[NeighborConfKey]NeighborData)
	intfConfEnt.NeighCreateCh = make(chan NeighCreateMsg)
	intfConfEnt.NeighChangeCh = make(chan NeighChangeMsg)
	intfConfEnt.NbrStateChangeCh = make(chan NbrStateChangeMsg)
	intfConfEnt.NbrFullStateCh = make(chan NbrFullStateMsg)

	intfConfEnt.rxHdl.RecvPcapHdl, err = server.initRxPkts(intfConfEnt.IfName, intfConfEnt.IpAddr)
	if err != nil {
		server.logger.Err("Error initializing Rx Pkt")
		return false, errors.New(fmt.Sprintln("Error initializing Rx Pkt", err))
	}
	intfConfEnt.rxHdl.PktRecvCtrlCh = make(chan bool)
	intfConfEnt.rxHdl.PktRecvCtrlReplyCh = make(chan bool)

	/*
		intfConfEnt.txHdl.SendPcapHdl, err = initTxPkts(initConfEnt.IfName)
		if err != nil {
			server.logger.Err("Error initializing Tx Pkt")
			return false, errors.New("Error initializing Tx Pkt", err)
		}

	*/
	intfConfEnt.LsaCount = 0

	server.IntfConfMap[intfConfKey] = intfConfEnt
	areaEnt.IntfMap[intfConfKey] = true
	server.AreaConfMap[cfg.AreaId] = areaEnt
	if server.globalData.AdminState == true &&
		cfg.AdminState == true {
		//StartSendAndRecvPkts : Start Tx and Rx and IntfFSM
		server.StartSendAndRecvPkts(intfConfKey)
		//TODO
		if len(areaEnt.IntfMap) == 1 {
			//Generate Router LSA
		} else {
			//Update Route LSA
		}
	}

	return true, nil
}

func (server *OSPFV2Server) deleteIntf(cfg *objects.Ospfv2Intf) (bool, error) {
	server.logger.Info("Intf configuration delete")
	intfConfKey := IntfConfKey{
		IpAddr:  cfg.IpAddress,
		IntfIdx: cfg.AddressLessIfIdx,
	}
	intfConfEnt, exist := server.IntfConfMap[intfConfKey]
	if !exist {
		server.logger.Err("Ospf Interface configuration doesnot exist")
		return false, errors.New("Ospf Interface configuration doesnot exist")
	}

	server.logger.Info("Intf Conf Ent", intfConfEnt)

	// TODO
	// Stop Interface FSM
	// Update Router LSA
	// Mark Network LSA to Max Age if DR
	// Declare all neighbors dead
	delete(server.IntfConfMap, intfConfKey)
	return true, nil
}

func (server *OSPFV2Server) getIntfState(ipAddr, addressLessIfIdx uint32) (*objects.Ospfv2IntfState, error) {
	var retObj objects.Ospfv2IntfState
	server.logger.Info("ipAddr:", ipAddr, "addressLessIfIdx:", addressLessIfIdx, server.IntfConfMap)
	return &retObj, nil
}

func (server *OSPFV2Server) getBulkIntfState(fromIdx, cnt int) (*objects.Ospfv2IntfStateGetInfo, error) {
	var retObj objects.Ospfv2IntfStateGetInfo
	server.logger.Info(server.IntfConfMap)
	return &retObj, nil
}
