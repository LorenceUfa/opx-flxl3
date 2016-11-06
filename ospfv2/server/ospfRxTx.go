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
	"l3/ospfv2/objects"
	"time"
)

func (server OSPFV2Server) StartSendAndRecvPkts(intfConfKey IntfConfKey) {
	ent, _ := server.IntfConfMap[intfConfKey]
	helloInterval := time.Duration(ent.HelloInterval) * time.Second
	ent.HelloIntervalTicker = time.NewTicker(helloInterval)
	if ent.Type == objects.INTF_TYPE_BROADCAST {
		waitTime := time.Duration(ent.RtrDeadInterval) * time.Second
		ent.WaitTimer = time.NewTimer(waitTime)
	}
	if ent.Type == objects.INTF_TYPE_BROADCAST {
		ent.State = objects.INTF_FSM_STATE_WAITING
	} else if ent.Type == objects.INTF_TYPE_POINT2POINT {
		ent.State = objects.INTF_FSM_STATE_P2P
	}
	server.IntfConfMap[intfConfKey] = ent
	server.logger.Info("Start Ospf Tx Pkt")
	//server.initTxPkt(intfConfKey)
	//server.StartOspfTxPkt(intfConfKey)
	server.logger.Info("Start Ospf Intf FSM")
	//go server.StartOspfIntfFSM(intfConfKey)
	server.logger.Info("Start Ospf Rx Pkt")
	go server.StartOspfRecvPkts(intfConfKey)
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
