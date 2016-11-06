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
	//"encoding/binary"
	//"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	//"l3/ospfv2/objects"
	//"net"
	"time"
)

type ProcessOspfPktRecvStruct struct {
	processOspfRecvPktCh       chan gopacket.Packet
	processOspfRecvCtrlCh      chan bool
	processOspfRecvCtrlReplyCh chan bool
	intfConfKey                IntfConfKey
}

func (server *OSPFV2Server) processOspfPkt(packet gopacket.Packet, intfConfKey IntfConfKey) {

}

func (server *OSPFV2Server) ProcessOspfRecvPkt(processRecvPkt ProcessOspfPktRecvStruct) {
	for {
		select {
		case packet := <-processRecvPkt.processOspfRecvPktCh:
			server.processOspfPkt(packet, processRecvPkt.intfConfKey)
		case _ = <-processRecvPkt.processOspfRecvCtrlCh:
			server.logger.Info("Stopping ProcessOspfRecvPkt")
			processRecvPkt.processOspfRecvCtrlReplyCh <- true
			return
		}
	}
}

func (server *OSPFV2Server) StartOspfRecvPkts(key IntfConfKey) {
	processOspfRecvCtrlCh := make(chan bool)
	processOspfRecvCtrlReplyCh := make(chan bool)
	processOspfRecvPktCh := make(chan gopacket.Packet, 1000)
	processRecvPkt := ProcessOspfPktRecvStruct{
		processOspfRecvPktCh:       processOspfRecvPktCh,
		processOspfRecvCtrlCh:      processOspfRecvCtrlCh,
		processOspfRecvCtrlReplyCh: processOspfRecvCtrlReplyCh,
		intfConfKey:                key,
	}
	ent, _ := server.IntfConfMap[key]
	go server.ProcessOspfRecvPkt(processRecvPkt)
	handle := ent.rxHdl.RecvPcapHdl
	recv := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := recv.Packets()
	for {
		select {
		case packet, ok := <-in:
			if ok {
				ipLayer := packet.Layer(layers.LayerTypeIPv4)
				if ipLayer == nil {
					server.logger.Err("Not an IP packet")
					continue
				}

				ipPkt := ipLayer.(*layers.IPv4)
				if ipPkt.Protocol == layers.IPProtocol(OSPF_PROTO_ID) {
					processRecvPkt.processOspfRecvPktCh <- packet
				}
			}
		case _ = <-ent.rxHdl.PktRecvCtrlCh:
			server.logger.Info("Stopping the Recv Ospf packet thread")
			processRecvPkt.processOspfRecvCtrlCh <- true
			_ = processRecvPkt.processOspfRecvCtrlReplyCh
			ent.rxHdl.PktRecvCtrlReplyCh <- true
			return
		}
	}
}

func (server *OSPFV2Server) StopOspfRecvPkts(key IntfConfKey) {
	intfEnt, _ := server.IntfConfMap[key]
	intfEnt.rxHdl.PktRecvCtrlCh <- true
	cnt := 0
	for {
		select {
		case _ = <-intfEnt.rxHdl.PktRecvCtrlReplyCh:
			server.logger.Info("Stopped Recv Pkt thread")
			return
		default:
			time.Sleep(time.Duration(10) * time.Millisecond)
			cnt = cnt + 1
			if cnt == 100 {
				server.logger.Err("Unable to stop the Rx thread")
				return
			}
		}
	}
}

func (server *OSPFV2Server) initRxPkts(ifName string, ipAddr uint32) (*pcap.Handle, error) {
	recvHdl, err := pcap.OpenLive(ifName, snapshotLen, promiscuous, pcapTimeout)
	if err != nil {
		server.logger.Err("Error opening recv pcap handler", ifName)
		return nil, err
	}
	ip := convertUint32ToDotNotation(ipAddr)
	filter := fmt.Sprintf("proto ospf and not src host %s", ip)
	server.logger.Info("Filter:", filter)
	err = recvHdl.SetBPFFilter(filter)
	if err != nil {
		server.logger.Err("Unable to set filter on", ifName)
		return nil, err
	}
	return recvHdl, nil
}
