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
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package server

import (
	"errors"
	"fmt"
	"l3/ospfv2/objects"
	"time"
)

func (server OSPFV2Server) StartSendAndRecvPkts(intfConfKey IntfConfKey) error {
	err := server.initRxTxPkts(intfConfKey)
	if err != nil {
		return err
	}

	ent, _ := server.IntfConfMap[intfConfKey]
	helloInterval := time.Duration(ent.HelloInterval) * time.Second
	ent.HelloIntervalTicker = time.NewTicker(helloInterval)
	if ent.Type == objects.INTF_TYPE_BROADCAST {
		waitTime := time.Duration(ent.RtrDeadInterval) * time.Second
		ent.WaitTimer = time.NewTimer(waitTime)
	}
	if ent.Type == objects.INTF_TYPE_BROADCAST {
		ent.FSMState = objects.INTF_FSM_STATE_WAITING
	} else if ent.Type == objects.INTF_TYPE_POINT2POINT {
		ent.FSMState = objects.INTF_FSM_STATE_P2P
	}
	server.IntfConfMap[intfConfKey] = ent
	server.logger.Info("Start Ospf Intf FSM")
	go server.StartOspfIntfFSM(intfConfKey)
	server.logger.Info("Start Ospf Rx Pkt")
	go server.StartOspfRecvPkts(intfConfKey)
	return nil
}

func (server *OSPFV2Server) StopSendAndRecvPkts(intfConfKey IntfConfKey) {
	server.StopOspfRecvPkts(intfConfKey)
	server.StopOspfIntfFSM(intfConfKey)
	ent, _ := server.IntfConfMap[intfConfKey]
	ent.FSMState = objects.INTF_FSM_STATE_DOWN
	if ent.Type == objects.INTF_TYPE_BROADCAST {
		ent.WaitTimer.Stop()
		ent.WaitTimer = nil
	}
	ent.HelloIntervalTicker.Stop()
	ent.HelloIntervalTicker = nil
	server.IntfConfMap[intfConfKey] = ent
	server.deinitRxTxPkts(intfConfKey)
}

func (server *OSPFV2Server) initRxTxPkts(intfConfKey IntfConfKey) error {
	var err error
	intfConfEnt, _ := server.IntfConfMap[intfConfKey]
	intfConfEnt.rxHdl.RecvPcapHdl, err = server.initRxPkts(intfConfEnt.IfName, intfConfEnt.IpAddr)
	if err != nil {
		server.logger.Err("Error initializing Rx Pkt")
		return errors.New(fmt.Sprintln("Error initializing Rx Pkt", err))
	}
	intfConfEnt.rxHdl.PktRecvCtrlCh = make(chan bool)
	intfConfEnt.rxHdl.PktRecvCtrlReplyCh = make(chan bool)

	intfConfEnt.txHdl.SendPcapHdl, err = server.initTxPkts(intfConfEnt.IfName)
	if err != nil {
		server.logger.Err("Error initializing Tx Pkt")
		return errors.New(fmt.Sprintln("Error initializing Tx Pkt", err))
	}
	server.IntfConfMap[intfConfKey] = intfConfEnt
	return nil
}

func (server *OSPFV2Server) deinitRxTxPkts(intfConfKey IntfConfKey) {
	intfConfEnt, exist := server.IntfConfMap[intfConfKey]
	if !exist {
		return
	}
	if intfConfEnt.rxHdl.RecvPcapHdl != nil {
		intfConfEnt.rxHdl.RecvPcapHdl.Close()
		intfConfEnt.rxHdl.PktRecvCtrlCh = nil
		intfConfEnt.rxHdl.PktRecvCtrlReplyCh = nil
	}
	intfConfEnt.txHdl.SendMutex.Lock()
	if intfConfEnt.txHdl.SendPcapHdl != nil {
		intfConfEnt.txHdl.SendPcapHdl.Close()
	}
	server.IntfConfMap[intfConfKey] = intfConfEnt
}

func (server *OSPFV2Server) StopAllIntfFSM() {
	for intfConfKey, intfConfEnt := range server.IntfConfMap {
		if intfConfEnt.FSMState != objects.INTF_FSM_STATE_DOWN {
			server.StopSendAndRecvPkts(intfConfKey)
		}
	}
}

func (server *OSPFV2Server) StartAllIntfFSM() {
	for intfConfKey, intfConfEnt := range server.IntfConfMap {
		if intfConfEnt.AdminState == true &&
			intfConfEnt.OperState == true {
			err := server.StartSendAndRecvPkts(intfConfKey)
			if err != nil {
				server.logger.Err("Error:", err)
			}
		}
	}
}

func (server *OSPFV2Server) StopAreaIntfFSM(areaId uint32) {
	areaEnt, _ := server.AreaConfMap[areaId]

	for intfConfKey, _ := range areaEnt.IntfMap {
		intfConfEnt, exist := server.IntfConfMap[intfConfKey]
		if !exist {
			server.logger.Err("IntfConfMap and AreaConfMap out of sync")
			continue
		}
		if intfConfEnt.FSMState != objects.INTF_FSM_STATE_DOWN {
			server.StopSendAndRecvPkts(intfConfKey)
		}
	}
}

func (server *OSPFV2Server) StartAreaIntfFSM(areaId uint32) {
	areaEnt, _ := server.AreaConfMap[areaId]

	for intfConfKey, _ := range areaEnt.IntfMap {
		intfConfEnt, exist := server.IntfConfMap[intfConfKey]
		if !exist {
			server.logger.Err("IntfConfMap and AreaConfMap out of sync")
			continue
		}
		if intfConfEnt.AdminState == true &&
			intfConfEnt.OperState == true {
			err := server.StartSendAndRecvPkts(intfConfKey)
			if err != nil {
				server.logger.Err("Error:", err)
			}
		}
	}
}

/*
func (server *OSPFServer) StartSendRecvPkts(intfConfKey IntfConfKey) {
        ent, _ := server.IntfConfMap[intfConfKey]
        server.updateIntfTxMap(intfConfKey, config.Intf_Up, ent.IfName)
        helloInterval := time.Duration(ent.IfHelloInterval) * time.Second
        ent.HelloIntervalTicker = time.NewTicker(helloInterval)
        if ent.IfType == config.Broadcast {
                waitTime := time.Duration(ent.IfRtrDeadInterval) * time.Second
                ent.WaitTimer = time.NewTimer(waitTime)
        }
        // rtrDeadInterval := time.Duration(ent.IfRtrDeadInterval * time.Second)
        ent.NeighborMap = make(map[NeighborConfKey]NeighborData)
        ent.IfEvents = ent.IfEvents + 1
        if ent.IfType == config.Broadcast {
                ent.IfFSMState = config.Waiting
        } else if ent.IfType == config.NumberedP2P || ent.IfType == config.UnnumberedP2P {
                ent.IfFSMState = config.P2P
        }
        server.IntfConfMap[intfConfKey] = ent
        server.logger.Info("Start Sending Hello Pkt")
        go server.StartOspfIntfFSM(intfConfKey)
        server.logger.Info("Start Receiving Hello Pkt")
        go server.StartOspfRecvPkts(intfConfKey)
}
*/
