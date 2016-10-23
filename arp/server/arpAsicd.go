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
	"utils/commonDefs"
)

type AsicdMsgType uint8

const (
	CreateAsicdEntry AsicdMsgType = 1
	DeleteAsicdEntry AsicdMsgType = 2
)

type AsicdMsg struct {
	MsgType AsicdMsgType
	IpAddr  string
	MacAddr string
	VlanId  int32
	IfIdx   int32
}

func (server *ARPServer) processAsicdNotification(msg commonDefs.AsicdNotifyMsg) {
	switch msg.(type) {
	case commonDefs.L2IntfStateNotifyMsg:
		l2Msg := msg.(commonDefs.L2IntfStateNotifyMsg)
		server.dumpInfra()
		server.logger.Info("L2IntfStateNotifyMsg:", l2Msg)
		server.processL2StateChange(l2Msg)
		server.dumpInfra()
	case commonDefs.IPv4L3IntfStateNotifyMsg:
		l3Msg := msg.(commonDefs.IPv4L3IntfStateNotifyMsg)
		server.dumpInfra()
		server.logger.Info("IPv4L3IntfStateNotifyMsg:", l3Msg)
		server.processIPv4L3StateChange(l3Msg)
		server.dumpInfra()
	case commonDefs.VlanNotifyMsg:
		vlanMsg := msg.(commonDefs.VlanNotifyMsg)
		server.dumpInfra()
		server.logger.Info("VlanNotifyMsg:", vlanMsg)
		server.updateVlanInfra(vlanMsg)
		server.dumpInfra()
	case commonDefs.LagNotifyMsg:
		lagMsg := msg.(commonDefs.LagNotifyMsg)
		server.dumpInfra()
		server.logger.Info("LagNotifyMsg:", lagMsg)
		server.updateLagInfra(lagMsg)
		server.dumpInfra()
	case commonDefs.IPv4IntfNotifyMsg:
		ipv4Msg := msg.(commonDefs.IPv4IntfNotifyMsg)
		server.dumpInfra()
		server.logger.Info("IPv4IntfNotifyMsg:", ipv4Msg)
		server.updateIPv4Infra(ipv4Msg)
		server.dumpInfra()
	case commonDefs.IPv4NbrMacMoveNotifyMsg:
		macMoveMsg := msg.(commonDefs.IPv4NbrMacMoveNotifyMsg)
		server.dumpInfra()
		server.processIPv4NbrMacMove(macMoveMsg)
		server.dumpInfra()
	}
}

func (server *ARPServer) dumpInfra() {
	server.dumpL3IntfProp()
	server.dumpVlanProp()
	//dumpLagProp()
	server.dumpPortProp()
}

func (server *ARPServer) dumpL3IntfProp() {
	server.logger.Info("==================================================")
	server.logger.Info("L3 Interface Property Map:")
	for l3IfIndex, l3Ent := range server.l3IntfPropMap {
		server.logger.Info("L3 IfIndex:", l3IfIndex, "IpAddr:", l3Ent.IpAddr, "Netmask:", l3Ent.Netmask, "IfName:", l3Ent.IfName)
	}
	server.logger.Info("==================================================")
}

func (server *ARPServer) dumpVlanProp() {
	server.logger.Info("==================================================")
	server.logger.Info("Vlan Property Map:")
	for vlanIfIdx, vlanEnt := range server.vlanPropMap {
		server.logger.Info("Vlan IfIdx:", vlanIfIdx, "Vlan Name:", vlanEnt.IfName)
		server.logger.Info("Untagged IfIndex Map:")
		for uIfIdx, _ := range vlanEnt.UntagIfIdxMap {
			server.logger.Info(uIfIdx)
		}
		server.logger.Info("Tagged IfIndex Map:")
		for tIfIdx, _ := range vlanEnt.TagIfIdxMap {
			server.logger.Info(tIfIdx)
		}
	}
	server.logger.Info("==================================================")
}

func (server *ARPServer) dumpPortProp() {
	server.logger.Info("==================================================")
	server.logger.Info(server.portPropMap[0])
	server.logger.Info(server.portPropMap[1])
	server.logger.Info(server.portPropMap[2])
	server.logger.Info(server.portPropMap[3])
	server.logger.Info(server.portPropMap[4])
	server.logger.Info("==================================================")
}

func (server *ARPServer) processAsicdMsg(msg AsicdMsg) error {
	switch msg.MsgType {
	case CreateAsicdEntry:
		_, err := server.AsicdPlugin.CreateIPv4Neighbor(msg.IpAddr, msg.MacAddr, msg.VlanId, msg.IfIdx)
		if err != nil {
			server.logger.Err("Asicd Create IPv4 Neighbor failed for IpAddr:", msg.IpAddr, "VlanId:", msg.VlanId, "IfIdx:", msg.IfIdx, "err:", err)
			return err
		}
	case DeleteAsicdEntry:
		_, err := server.AsicdPlugin.DeleteIPv4Neighbor(msg.IpAddr)
		if err != nil {
			server.logger.Err("Asicd was unable to delete neigbhor entry for", msg.IpAddr, "err:", err)
			return err
		}
	default:
		err := errors.New("Invalid Asicd Msg Type")
		return err
	}
	return nil
}
