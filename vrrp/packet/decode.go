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
package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"l3/vrrp/common"
	"l3/vrrp/debug"
	"syscall"
)

/*
	0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                    IPv4 Fields or IPv6 Fields                 |
	...                                                             ...
	|                                                               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|Version| Type  | Virtual Rtr ID|   Priority    |Count IPvX Addr|
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|(rsvd) |     Max Adver Int     |          Checksum             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                                                               |
	+                                                               +
	|                       IPvX Address(es)                        |
	+                                                               +
	+                                                               +
	+                                                               +
	+                                                               +
	|                                                               |
	+                                                               +
	|                                                               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

func (p *PacketInfo) ValidateHeader(hdr *Header, layerContent []byte) error {
	if hdr.Version != common.VERSION2 && hdr.Version != common.VERSION3 {
		return errors.New(VRRP_INCORRECT_VERSION)
	}
	// Set Checksum to 0 for verifying checksum
	binary.BigEndian.PutUint16(layerContent[6:8], 0)
	// Verify checksum
	chksum := computeChecksum(hdr.Version, layerContent)
	if chksum != hdr.CheckSum {
		debug.Logger.Err(chksum, "!=", hdr.CheckSum)
		return errors.New(fmt.Sprintln(VRRP_CHECKSUM_ERR, "computed CheckSum:", chksum, "received chksum in pkt:", hdr.CheckSum))
	}

	// Verify VRRP fields
	if hdr.CountIPAddr == 0 || hdr.MaxAdverInt == 0 || hdr.Type == 0 {
		return errors.New(VRRP_INCORRECT_FIELDS)
	}
	return nil
}

func DecodeHeader(data []byte) *Header {
	var hdr Header
	hdr.Version = uint8(data[0]) >> 4
	hdr.Type = uint8(data[0]) & 0x0F
	hdr.VirtualRtrId = data[1]
	hdr.Priority = data[2]
	hdr.CountIPAddr = data[3]
	rsvdAdver := binary.BigEndian.Uint16(data[4:6])
	hdr.Rsvd = uint8(rsvdAdver >> 13)
	hdr.MaxAdverInt = rsvdAdver & 0x1FFF
	hdr.CheckSum = binary.BigEndian.Uint16(data[6:8])
	var ipvxAddrlength int
	if VRRP_HEADER_SIZE_EXCLUDING_IPVX > len(data) {
		ipvxAddrlength = VRRP_HEADER_SIZE_EXCLUDING_IPVX - len(data)
	} else {
		ipvxAddrlength = len(data) - VRRP_HEADER_SIZE_EXCLUDING_IPVX
	}
	baseIpByte := 8
	switch hdr.Version {
	case common.VERSION2:
		for i := 0; i < int(hdr.CountIPAddr); i++ {
			hdr.IpAddr = append(hdr.IpAddr, data[baseIpByte:(baseIpByte+4)])
			baseIpByte += 4
		}
	case common.VERSION3:
		var ipType int
		if int(hdr.CountIPAddr)*16 == ipvxAddrlength {
			ipType = syscall.AF_INET6
		} else if int(hdr.CountIPAddr)*4 == ipvxAddrlength {
			ipType = syscall.AF_INET
		}
		for i := 0; i < int(hdr.CountIPAddr); i++ {
			if ipType == syscall.AF_INET {
				hdr.IpAddr = append(hdr.IpAddr, data[baseIpByte:(baseIpByte+4)])
				baseIpByte += 4
			} else if ipType == syscall.AF_INET6 {
				hdr.IpAddr = append(hdr.IpAddr, data[baseIpByte:(baseIpByte+16)])
				baseIpByte += 16
			}
		}
	}
	return &hdr
}

func (p *PacketInfo) Decode(pkt gopacket.Packet, ipType int) *PacketInfo {
	// Check dmac address from the inPacket and if it is same discard the pkt
	ethLayer := pkt.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		debug.Logger.Err("Not an eth packet?")
		return nil
	}
	eth := ethLayer.(*layers.Ethernet)
	var ipLayer gopacket.Layer
	var dstIp string
	var srcIp string
	switch eth.EthernetType {
	case layers.EthernetTypeIPv4:
		// Get Entire IP layer Info
		ipLayer = pkt.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			debug.Logger.Err("Not an ip packet?")
			return nil
		}
		// Get Ip Hdr and start doing basic check according to RFC
		ipHdr := ipLayer.(*layers.IPv4)
		if ipHdr.TTL != VRRP_TTL {
			debug.Logger.Err("ttl should be 255 instead of", ipHdr.TTL,
				"dropping packet from", ipHdr.SrcIP.String())
			return nil
		}
		dstIp = ipHdr.DstIP.String()
		srcIp = ipHdr.SrcIP.String()
	case layers.EthernetTypeIPv6:
		// Get Entire IP layer Info
		ipLayer = pkt.Layer(layers.LayerTypeIPv6)
		if ipLayer == nil {
			debug.Logger.Err("Not an ip packet?")
			return nil
		}
		ipHdr := ipLayer.(*layers.IPv6)
		dstIp = ipHdr.DstIP.String()
		srcIp = ipHdr.SrcIP.String()
	}
	// Get Payload as checks are succesful
	ipPayload := ipLayer.LayerPayload()
	if ipPayload == nil {
		debug.Logger.Err("No payload for ip packet")
		return nil
	}
	// Get VRRP header from IP Payload
	hdr := DecodeHeader(ipPayload)
	// Do Basic Vrrp Header Check
	if err := p.ValidateHeader(hdr, ipPayload); err != nil {
		debug.Logger.Err(err.Error(), ". Dropping received packet from", srcIp)
		return nil
	}
	pktInfo := &PacketInfo{
		Hdr:    hdr,
		DstIp:  dstIp,
		IpAddr: srcIp,
		DstMac: (eth.DstMAC).String(),
		SrcMac: (eth.SrcMAC).String(),
	}
	return pktInfo
}
