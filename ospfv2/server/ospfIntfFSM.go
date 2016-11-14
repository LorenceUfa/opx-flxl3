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
	"l3/ospfv2/objects"
	"time"
)

func (server *OSPFV2Server) StopOspfIntfFSM(key IntfConfKey) {
	ent, _ := server.IntfConfMap[key]
	ent.FSMCtrlCh <- false
	cnt := 0
	for {
		select {
		case _ = <-ent.FSMCtrlReplyCh:
			server.logger.Info("Stopped Sending Hello Pkt")
			return
		default:
			time.Sleep(time.Duration(10) * time.Millisecond)
			cnt = cnt + 1
			if cnt == 100 {
				server.logger.Err("Unable to stop the Tx thread")
				return
			}
		}
	}
}

func (server *OSPFV2Server) StartOspfIntfFSM(key IntfConfKey) {
	ent, _ := server.IntfConfMap[key]
	if ent.Type == objects.INTF_TYPE_POINT2POINT {
		server.StartOspfP2PIntfFSM(key)
	} else if ent.Type == objects.INTF_TYPE_BROADCAST {
		server.StartOspfBroadcastIntfFSM(key)
	}
}

func (server *OSPFV2Server) StartOspfP2PIntfFSM(key IntfConfKey) {
	server.StartSendHelloPkt(key)
	for {
		ent, exist := server.IntfConfMap[key]
		if !exist {
			server.logger.Err("Interface does not exist", key)
			return
		}
		nbrDownMsgCh, _ := server.MessagingChData.NbrToIntfFSMChData.NbrDownMsgChMap[key]
		select {
		case <-ent.HelloIntervalTicker.C:
			server.StartSendHelloPkt(key)
		case createMsg := <-ent.NbrCreateCh:
			if createMsg.DRtrIpAddr != 0 ||
				createMsg.BDRtrIpAddr != 0 {
				server.logger.Err("DR or BDR is non zero")
				continue
			}
			nbrKey := createMsg.NbrKey
			nbrEntry, exist := ent.NbrMap[nbrKey]
			if !exist {
				nbrEntry.TwoWayStatus = createMsg.TwoWayStatus
				nbrEntry.RtrPrio = createMsg.RtrPrio
				nbrEntry.DRtrIpAddr = createMsg.DRtrIpAddr
				nbrEntry.BDRtrIpAddr = createMsg.BDRtrIpAddr
				nbrEntry.NbrIpAddr = createMsg.NbrIP
				nbrEntry.RtrId = createMsg.RouterId
				ent.NbrMap[nbrKey] = nbrEntry
				server.IntfConfMap[key] = ent
			}
		case changeMsg := <-ent.NbrChangeCh:
			if changeMsg.DRtrIpAddr != 0 ||
				changeMsg.BDRtrIpAddr != 0 {
				server.logger.Err("DR or BDR is non zero")
				continue
			}
			nbrKey := changeMsg.NbrKey
			nbrEntry, exist := ent.NbrMap[nbrKey]
			if exist {
				nbrEntry.TwoWayStatus = changeMsg.TwoWayStatus
				nbrEntry.RtrPrio = changeMsg.RtrPrio
				nbrEntry.NbrIpAddr = changeMsg.NbrIP
				nbrEntry.DRtrIpAddr = changeMsg.DRtrIpAddr
				nbrEntry.BDRtrIpAddr = changeMsg.BDRtrIpAddr
				nbrEntry.RtrId = changeMsg.RouterId
				ent.NbrMap[nbrKey] = nbrEntry
				server.IntfConfMap[key] = ent
			} else {
				server.logger.Err("Nbr entry does not exists", nbrKey)
			}
		case downMsg := <-nbrDownMsgCh:
			// Only when Nbr Went Down from TwoWayStatus
			server.logger.Info("Recev Nbr State Change message", downMsg)
			server.processNbrDownEvent(downMsg, key, true)
		case _ = <-ent.FSMCtrlCh:
			server.StopSendHelloPkt(key)
			ent.FSMCtrlReplyCh <- false
			return
		}
	}
}

func (server *OSPFV2Server) StartOspfBroadcastIntfFSM(key IntfConfKey) {
	server.StartSendHelloPkt(key)
	for {
		ent, exist := server.IntfConfMap[key]
		if !exist {
			server.logger.Err("Intf conf doesnot exist", key)
			return
		}
		nbrDownMsgCh, _ := server.MessagingChData.NbrToIntfFSMChData.NbrDownMsgChMap[key]
		select {
		case <-ent.HelloIntervalTicker.C:
			server.StartSendHelloPkt(key)
		case <-ent.WaitTimer.C:
			server.logger.Info("Wait timer expired")
			server.ElectBDRAndDR(key)
		case msg := <-ent.BackupSeenCh:
			server.logger.Info("Transit to action state because of backup seen", msg)
			server.ElectBDRAndDR(key)
		case createMsg := <-ent.NbrCreateCh:
			nbrKey := createMsg.NbrKey
			nbrEntry, exist := ent.NbrMap[nbrKey]
			if !exist {
				nbrEntry.TwoWayStatus = createMsg.TwoWayStatus
				nbrEntry.RtrPrio = createMsg.RtrPrio
				nbrEntry.DRtrIpAddr = createMsg.DRtrIpAddr
				nbrEntry.BDRtrIpAddr = createMsg.BDRtrIpAddr
				nbrEntry.NbrIpAddr = createMsg.NbrIP
				nbrEntry.RtrId = createMsg.RouterId
				ent.NbrMap[nbrKey] = nbrEntry
				server.IntfConfMap[key] = ent
				if createMsg.TwoWayStatus == true &&
					ent.FSMState > objects.INTF_FSM_STATE_WAITING {
					server.ElectBDRAndDR(key)
				}
			}
		case changeMsg := <-ent.NbrChangeCh:
			nbrKey := changeMsg.NbrKey
			nbrEntry, exist := ent.NbrMap[nbrKey]
			if exist {
				//rtrId := changeMsg.RouterId
				NbrIP := changeMsg.NbrIP
				oldRtrPrio := nbrEntry.RtrPrio
				oldDRtrIpAddr := nbrEntry.DRtrIpAddr
				oldBDRtrIpAddr := nbrEntry.BDRtrIpAddr
				oldTwoWayStatus := nbrEntry.TwoWayStatus
				newDRtrIpAddr := changeMsg.DRtrIpAddr
				newBDRtrIpAddr := changeMsg.BDRtrIpAddr
				nbrEntry.NbrIpAddr = changeMsg.NbrIP
				nbrEntry.TwoWayStatus = changeMsg.TwoWayStatus
				nbrEntry.RtrPrio = changeMsg.RtrPrio
				nbrEntry.DRtrIpAddr = changeMsg.DRtrIpAddr
				nbrEntry.BDRtrIpAddr = changeMsg.BDRtrIpAddr
				nbrEntry.RtrId = changeMsg.RouterId
				ent.NbrMap[nbrKey] = nbrEntry
				server.IntfConfMap[key] = ent
				if ent.FSMState > objects.INTF_FSM_STATE_WAITING {
					// RFC2328 Section 9.2 (Nbr Change Event)
					if (oldDRtrIpAddr == NbrIP && newDRtrIpAddr != NbrIP && oldTwoWayStatus == true) ||
						(oldDRtrIpAddr != NbrIP && newDRtrIpAddr == NbrIP && oldTwoWayStatus == true) ||
						(oldBDRtrIpAddr == NbrIP && newBDRtrIpAddr != NbrIP && oldTwoWayStatus == true) ||
						(oldBDRtrIpAddr != NbrIP && newBDRtrIpAddr == NbrIP && oldTwoWayStatus == true) ||
						(oldTwoWayStatus != changeMsg.TwoWayStatus) ||
						(oldRtrPrio != changeMsg.RtrPrio && oldTwoWayStatus == true) {

						// Update Nbr and Re-elect BDR And DR
						server.ElectBDRAndDR(key)
					}
				}
			}
		case downMsg := <-nbrDownMsgCh:
			// Only when Nbr Went Down from TwoWayStatus
			server.logger.Info("Recev Nbr State Change message", downMsg)
			server.processNbrDownEvent(downMsg, key, false)
		case _ = <-ent.FSMCtrlCh:
			server.StopSendHelloPkt(key)
			ent.FSMCtrlReplyCh <- false
			return
		}
	}
}

func (server *OSPFV2Server) processNbrDownEvent(msg NbrDownMsg,
	key IntfConfKey, p2p bool) {
	ent, _ := server.IntfConfMap[key]
	_, exist := ent.NbrMap[msg.NbrKey]
	if exist {
		delete(ent.NbrMap, msg.NbrKey)
		server.logger.Info("Deleting", msg.NbrKey)
		server.IntfConfMap[key] = ent
		if p2p == false {
			if ent.FSMState > objects.INTF_FSM_STATE_WAITING {
				// RFC2328 Section 9.2 (Nbr Change Event)
				server.logger.Info("deleting nbr.")
				server.ElectBDRAndDR(key)
			}
		}
	}
}

func (server *OSPFV2Server) ElectBDR(key IntfConfKey) (uint32, uint32) {
	ent, _ := server.IntfConfMap[key]
	electedBDRIpAddr := uint32(0)
	var electedRtrPrio uint8
	var electedRtrId uint32
	var MaxRtrPrio uint8
	var RtrIdWithMaxPrio uint32
	var NbrIPWithMaxPrio uint32

	for _, nbrEntry := range ent.NbrMap {
		if nbrEntry.TwoWayStatus == true &&
			nbrEntry.RtrPrio > 0 &&
			nbrEntry.NbrIpAddr != 0 {
			tempDRIpAddr := nbrEntry.DRtrIpAddr
			if tempDRIpAddr == nbrEntry.NbrIpAddr {
				continue
			}
			tempBDRIpAddr := nbrEntry.BDRtrIpAddr
			if tempBDRIpAddr == nbrEntry.NbrIpAddr {
				if nbrEntry.RtrPrio > electedRtrPrio {
					electedRtrPrio = nbrEntry.RtrPrio
					electedRtrId = nbrEntry.RtrId
					electedBDRIpAddr = nbrEntry.BDRtrIpAddr
				} else if nbrEntry.RtrPrio == electedRtrPrio {
					if electedRtrId < nbrEntry.RtrId {
						electedRtrPrio = nbrEntry.RtrPrio
						electedRtrId = nbrEntry.RtrId
						electedBDRIpAddr = nbrEntry.BDRtrIpAddr
					}
				}
			}
			if MaxRtrPrio < nbrEntry.RtrPrio {
				MaxRtrPrio = nbrEntry.RtrPrio
				RtrIdWithMaxPrio = nbrEntry.RtrId
				NbrIPWithMaxPrio = nbrEntry.NbrIpAddr
			} else if MaxRtrPrio == nbrEntry.RtrPrio {
				if RtrIdWithMaxPrio < nbrEntry.RtrId {
					MaxRtrPrio = nbrEntry.RtrPrio
					RtrIdWithMaxPrio = nbrEntry.RtrId
					NbrIPWithMaxPrio = nbrEntry.NbrIpAddr
				}
			}
		}
	}

	if ent.RtrPriority != 0 && ent.IpAddr != 0 {
		if ent.IpAddr != ent.DRIpAddr {
			if ent.IpAddr != ent.BDRIpAddr {
				rtrId := server.globalData.RouterId
				if ent.RtrPriority > electedRtrPrio {
					electedRtrPrio = ent.RtrPriority
					electedRtrId = rtrId
					electedBDRIpAddr = ent.IpAddr
				} else if ent.RtrPriority == electedRtrPrio {
					if electedRtrId < rtrId {
						electedRtrPrio = ent.RtrPriority
						electedRtrId = rtrId
						electedBDRIpAddr = ent.IpAddr
					}
				}
			}

			tempRtrId := server.globalData.RouterId
			if MaxRtrPrio < ent.RtrPriority {
				MaxRtrPrio = ent.RtrPriority
				NbrIPWithMaxPrio = ent.IpAddr
				RtrIdWithMaxPrio = tempRtrId
			} else if MaxRtrPrio == ent.RtrPriority {
				if RtrIdWithMaxPrio < tempRtrId {
					MaxRtrPrio = ent.RtrPriority
					NbrIPWithMaxPrio = ent.IpAddr
					RtrIdWithMaxPrio = tempRtrId
				}
			}

		}
	}
	if electedBDRIpAddr == 0 {
		electedBDRIpAddr = NbrIPWithMaxPrio
		electedRtrId = RtrIdWithMaxPrio
	}

	return electedBDRIpAddr, electedRtrId
}

func (server *OSPFV2Server) ElectDR(key IntfConfKey, electedBDRIpAddr uint32, electedBDRtrId uint32) (uint32, uint32) {
	ent, _ := server.IntfConfMap[key]
	electedDRIpAddr := uint32(0)
	var electedRtrPrio uint8
	var electedDRtrId uint32

	for _, nbrEntry := range ent.NbrMap {
		if nbrEntry.TwoWayStatus == true &&
			nbrEntry.RtrPrio > 0 &&
			nbrEntry.NbrIpAddr != 0 {
			tempDRIpAddr := nbrEntry.DRtrIpAddr
			if tempDRIpAddr == nbrEntry.NbrIpAddr {
				if nbrEntry.RtrPrio > electedRtrPrio {
					electedRtrPrio = nbrEntry.RtrPrio
					electedDRtrId = nbrEntry.RtrId
					electedDRIpAddr = nbrEntry.DRtrIpAddr
				} else if nbrEntry.RtrPrio == electedRtrPrio {
					if electedDRtrId < nbrEntry.RtrId {
						electedRtrPrio = nbrEntry.RtrPrio
						electedDRtrId = nbrEntry.RtrId
						electedDRIpAddr = nbrEntry.DRtrIpAddr
					}
				}
			}
		}
	}

	if ent.RtrPriority > 0 &&
		ent.IpAddr != 0 {
		if ent.IpAddr == ent.DRIpAddr {
			rtrId := server.globalData.RouterId
			if ent.RtrPriority > electedRtrPrio {
				electedRtrPrio = ent.RtrPriority
				electedDRtrId = rtrId
				electedDRIpAddr = ent.IpAddr
			} else if ent.RtrPriority == electedRtrPrio {
				if electedDRtrId < rtrId {
					electedRtrPrio = ent.RtrPriority
					electedDRtrId = rtrId
					electedDRIpAddr = ent.IpAddr
				}
			}
		}
	}

	if electedDRIpAddr == 0 {
		electedDRIpAddr = electedBDRIpAddr
		electedDRtrId = electedBDRtrId
	}
	return electedDRIpAddr, electedDRtrId
}

func (server *OSPFV2Server) ElectBDRAndDR(key IntfConfKey) {
	ent, _ := server.IntfConfMap[key]
	server.logger.Info("Election of BDR andDR", ent.FSMState)

	oldDRtrIpAddr := ent.DRIpAddr
	//oldBDRtrIpAddr := ent.BDRIpAddr
	oldDRtrId := ent.DRtrId
	//oldBDRtrId := ent.BDRtrId

	oldState := ent.FSMState
	var newState uint8

	electedBDRIpAddr, electedBDRtrId := server.ElectBDR(key)
	ent.BDRIpAddr = electedBDRIpAddr
	ent.BDRtrId = electedBDRtrId
	electedDRIpAddr, electedDRtrId := server.ElectDR(key, electedBDRIpAddr, electedBDRtrId)
	ent.DRIpAddr = electedDRIpAddr
	ent.DRtrId = electedDRtrId
	if ent.DRIpAddr == ent.IpAddr {
		newState = objects.INTF_FSM_STATE_DR
	} else if ent.BDRIpAddr == ent.IpAddr {
		newState = objects.INTF_FSM_STATE_BDR
	} else {
		newState = objects.INTF_FSM_STATE_OTHER_DR
	}

	ent.FSMState = newState
	server.IntfConfMap[key] = ent

	if newState != oldState &&
		!(newState == objects.INTF_FSM_STATE_OTHER_DR &&
			oldState < objects.INTF_FSM_STATE_OTHER_DR) {
		ent, _ = server.IntfConfMap[key]
		electedBDRIpAddr, electedBDRtrId = server.ElectBDR(key)
		ent.BDRIpAddr = electedBDRIpAddr
		ent.BDRtrId = electedBDRtrId
		electedDRIpAddr, electedDRtrId = server.ElectDR(key, electedBDRIpAddr, electedBDRtrId)
		ent.DRIpAddr = electedDRIpAddr
		ent.DRtrId = electedDRtrId
		if ent.DRIpAddr == ent.IpAddr {
			newState = objects.INTF_FSM_STATE_DR
		} else if ent.BDRIpAddr == ent.IpAddr {
			newState = objects.INTF_FSM_STATE_BDR
		} else {
			newState = objects.INTF_FSM_STATE_OTHER_DR
		}
		ent.FSMState = newState
		server.IntfConfMap[key] = ent
	}

	if oldDRtrId != ent.DRtrId || oldDRtrIpAddr != ent.DRIpAddr {
		server.ProcessNetworkDRChange(key, ent.AreaId, oldState, newState)
	}
}

func (server *OSPFV2Server) ProcessNetworkDRChange(key IntfConfKey, areaId uint32, oldState, newState uint8) {
	server.SendMsgToGenerateRouterLSA(areaId)
	server.SendNetworkDRChangeMsg(key, oldState, newState)
}
