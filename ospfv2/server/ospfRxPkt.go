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
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"l3/ospfv2/objects"
	"time"
)

type ProcessOspfPktRecvStruct struct {
	processOspfRecvPktCh       chan gopacket.Packet
	processOspfRecvCtrlCh      chan bool
	processOspfRecvCtrlReplyCh chan bool
	intfConfKey                IntfConfKey
}

func (server *OSPFV2Server) processOspfData(data []byte, ethHdrMd *EthHdrMetadata, ipHdrMd *IpHdrMetadata, ospfHdrMd *OspfHdrMetadata, key IntfConfKey) error {
	var err error = nil
	switch ospfHdrMd.PktType {
	case HelloType:
		err = server.processRxHelloPkt(data, ospfHdrMd, ipHdrMd, ethHdrMd, key)
	case DBDescriptionType:
		//err = server.ProcessRxDbdPkt(data, ospfHdrMd, ipHdrMd, key, ethHdrMd.srcMAC)
	case LSRequestType:
		//err = server.ProcessRxLSAReqPkt(data, ospfHdrMd, ipHdrMd, key)
	case LSUpdateType:
		//err = server.ProcessRxLsaUpdPkt(data, ospfHdrMd, ipHdrMd, key)
	case LSAckType:
		//err = server.ProcessRxLSAAckPkt(data, ospfHdrMd, ipHdrMd, key)
	default:
		err = errors.New("Invalid Ospf packet type")
	}
	return err
}

func (server *OSPFV2Server) processIPv4Layer(ipLayer gopacket.Layer, IpAddr uint32, ipHdrMd *IpHdrMetadata) error {
	ipLayerContents := ipLayer.LayerContents()
	ipChkSum := binary.BigEndian.Uint16(ipLayerContents[10:12])
	binary.BigEndian.PutUint16(ipLayerContents[10:], 0)

	csum := computeCheckSum(ipLayerContents)
	if csum != ipChkSum {
		err := errors.New("Incorrect IPv4 checksum, hence dicarding the packet")
		return err
	}

	ipPkt := ipLayer.(*layers.IPv4)
	ipHdrMd.SrcIP, _ = convertDotNotationToUint32(ipPkt.SrcIP.To4().String())
	ipHdrMd.DstIP, _ = convertDotNotationToUint32(ipPkt.DstIP.To4().String())
	if IpAddr == ipHdrMd.SrcIP {
		err := errors.New(fmt.Sprintln("locally generated pkt", ipPkt.SrcIP, "hence dicarding the packet"))
		return err
	}

	if IpAddr != ipHdrMd.DstIP &&
		ALLDROUTER != ipHdrMd.DstIP &&
		ALLSPFROUTER != ipHdrMd.DstIP {
		err := errors.New(fmt.Sprintln("Incorrect DstIP", ipPkt.DstIP, "hence dicarding the packet"))
		return err
	}

	if ipPkt.Protocol != layers.IPProtocol(OSPF_PROTO_ID) {
		err := errors.New(fmt.Sprintln("Incorrect ProtocolID", ipPkt.Protocol, "hence dicarding the packet"))
		return err
	}
	if ALLSPFROUTER == ipHdrMd.DstIP {
		ipHdrMd.DstIPType = AllSPFRouterType
	} else if ALLDROUTER == ipHdrMd.DstIP {
		ipHdrMd.DstIPType = AllDRouterType
	} else {
		ipHdrMd.DstIPType = NormalType
	}
	return nil
}

func (server *OSPFV2Server) processOspfHeader(ospfPkt []byte, key IntfConfKey, md *OspfHdrMetadata, ipHdrMd *IpHdrMetadata) error {
	if len(ospfPkt) < OSPF_HEADER_SIZE {
		err := errors.New("Invalid length of Ospf Header")
		return err
	}

	ent, exist := server.IntfConfMap[key]
	if !exist {
		err := errors.New("Dropped because of interface no more valid")
		return err
	}

	ospfHdr := NewOSPFHeader()

	decodeOspfHdr(ospfPkt, ospfHdr)

	if OSPF_VERSION_2 != ospfHdr.Ver {
		err := errors.New("Dropped because of Ospf Version not matching")
		return err
	}

	if ent.AreaId == ospfHdr.AreaId {
		if ent.Type != objects.INTF_TYPE_POINT2POINT {
			if (ent.IpAddr & ent.Netmask) != (ipHdrMd.SrcIP & ent.Netmask) {
				err := errors.New("Dropped because of Src IP is not in subnet and Area ID is matching")
				return err

			}
		}
	} else {
		// We don't support Virtual Link
		err := errors.New("Dropped because Area ID is not matching and we dont support Virtual links, so this should not happend")
		return err

	}

	if ipHdrMd.DstIPType == AllDRouterType {
		if ent.DRtrId != server.globalData.RouterId &&
			ent.BDRtrId != server.globalData.RouterId {
			err := errors.New("Dropped because we should not recv any pkt with ALLDROUTER as we are not DR or BDR")
			return err
		}
	}

	//OSPF Auth Type
	if ent.AuthType != ospfHdr.AuthType {
		err := errors.New("Dropped because of Router Id not matching")
		return err
	}

	//TODO: We don't support Authentication

	if ospfHdr.PktType != HelloType {
		if ent.Type == objects.INTF_TYPE_BROADCAST {
			nbrKey := NbrConfKey{
				NbrIdentity:         ipHdrMd.SrcIP,
				NbrAddressLessIfIdx: key.IntfIdx,
			}
			_, exist := ent.NbrMap[nbrKey]
			if !exist {
				err := errors.New("Adjacency not established with this nbr")
				return err
			}
		} else if ent.Type == objects.INTF_TYPE_POINT2POINT {
			nbrKey := NbrConfKey{
				NbrIdentity:         ospfHdr.RouterId,
				NbrAddressLessIfIdx: key.IntfIdx,
			}
			_, exist := ent.NbrMap[nbrKey]
			if !exist {
				err := errors.New("Adjacency not established with this nbr")
				return err
			}
		}
	}

	//OSPF Header CheckSum
	binary.BigEndian.PutUint16(ospfPkt[12:14], 0)
	copy(ospfPkt[16:OSPF_HEADER_SIZE], []byte{0, 0, 0, 0, 0, 0, 0, 0})
	csum := computeCheckSum(ospfPkt)
	if csum != ospfHdr.Chksum {
		err := errors.New("Dropped because of invalid checksum")
		return err
	}

	md.PktType = ospfHdr.PktType
	md.Pktlen = ospfHdr.Pktlen
	md.RouterId = ospfHdr.RouterId
	md.AreaId = ospfHdr.AreaId
	if ospfHdr.AreaId == 0 {
		md.Backbone = true
	} else {
		md.Backbone = false
	}

	return nil
}

func (server *OSPFV2Server) processOspfPkt(pkt gopacket.Packet, key IntfConfKey) {
	server.logger.Info("Recevied Ospf Packet")
	ent, exist := server.IntfConfMap[key]
	if !exist {
		server.logger.Err("Dropped because of interface no more valid")
		return
	}

	ethLayer := pkt.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		server.logger.Err("Not an Ethernet frame")
		return
	}
	eth := ethLayer.(*layers.Ethernet)

	ethHdrMd := NewEthHdrMetadata()
	ethHdrMd.SrcMAC = eth.SrcMAC

	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		server.logger.Err("Not an IP packet")
		return
	}

	ipHdrMd := NewIpHdrMetadata()
	err := server.processIPv4Layer(ipLayer, ent.IpAddr, ipHdrMd)
	if err != nil {
		server.logger.Err("Dropped because of IPv4 layer processing", err)
		return
	}

	ospfHdrMd := NewOspfHdrMetadata()
	ospfPkt := ipLayer.LayerPayload()
	err = server.processOspfHeader(ospfPkt, key, ospfHdrMd, ipHdrMd)
	if err != nil {
		server.logger.Err("Dropped because of Ospf Header processing", err)
		return
	}

	ospfData := ospfPkt[OSPF_HEADER_SIZE:]
	err = server.processOspfData(ospfData, ethHdrMd, ipHdrMd, ospfHdrMd, key)
	if err != nil {
		server.logger.Err("Dropped because of Ospf Header processing", err)
		return
	}
	return
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
