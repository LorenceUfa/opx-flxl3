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
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"l3/vrrp/debug"
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

type Header struct {
	Version       uint8
	Type          uint8
	VirtualRtrId  uint8
	Priority      uint8
	CountIPv4Addr uint8
	Rsvd          uint8
	MaxAdverInt   uint16
	CheckSum      uint16
	IPv4Addr      []net.IP
}

func DecodeHeader(data []byte) *Header {
	var hdr Header
	hdr.Version = uint8(data[0]) >> 4
	hdr.Type = uint8(data[0]) & 0x0F
	hdr.VirtualRtrId = data[1]
	hdr.Priority = data[2]
	hdr.CountIPv4Addr = data[3]
	rsvdAdver := binary.BigEndian.Uint16(data[4:6])
	hdr.Rsvd = uint8(rsvdAdver >> 13)
	hdr.MaxAdverInt = rsvdAdver & 0x1FFF
	hdr.CheckSum = binary.BigEndian.Uint16(data[6:8])
	baseIpByte := 8
	for i := 0; i < int(hdr.CountIPv4Addr); i++ {
		hdr.IPv4Addr = append(hdr.IPv4Addr, data[baseIpByte:(baseIpByte+4)])
		baseIpByte += 4
	}
	return &hdr
}

func computeChecksum(version uint8, content []byte) uint16 {
	var csum uint32
	var rv uint16
	if version == VRRP_VERSION2 {
		for i := 0; i < len(content); i += 2 {
			csum += uint32(content[i]) << 8
			csum += uint32(content[i+1])
		}
		rv = ^uint16((csum >> 16) + csum)
	} else if version == VRRP_VERSION3 {
		//@TODO: .....
	}

	return rv
}

func (p *PacketInfo) ValidateHeader(hdr *Header, layerContent []byte) error { //, key string) error {
	// @TODO: need to check for version 2 type...RFC requests to drop the pkt
	// but cisco uses version 2...
	if hdr.Version != VRRP_VERSION2 && hdr.Version != VRRP_VERSION3 {
		return errors.New(VRRP_INCORRECT_VERSION)
	}
	// Set Checksum to 0 for verifying checksum
	binary.BigEndian.PutUint16(layerContent[6:8], 0)
	// Verify checksum
	chksum := computeChecksum(hdr.Version, layerContent)
	if chksum != hdr.CheckSum {
		debug.Logger.Err(chksum, "!=", hdr.CheckSum)
		return errors.New(VRRP_CHECKSUM_ERR)
	}

	// Verify VRRP fields
	if hdr.CountIPv4Addr == 0 ||
		hdr.MaxAdverInt == 0 ||
		hdr.Type == 0 {
		return errors.New(VRRP_INCORRECT_FIELDS)
	}
	/*
		gblInfo := p.vrrpGblInfo[key]
		if gblInfo.IntfConfig.VirtualIPv4Addr == "" {
			for i := 0; i < int(hdr.CountIPv4Addr); i++ {
				/* If Virtual Ip is not configured then check whether the ip
				 * address of router/interface is not same as the received
				 * Virtual Ip Addr
	*/
	/*
				if gblInfo.IpAddr == hdr.IPv4Addr[i].String() {
					return errors.New(VRRP_SAME_OWNER)
				}
			}
		}

		if gblInfo.IntfConfig.VRID == 0 {
			return errors.New(VRRP_MISSING_VRID_CONFIG)
		}
	*/
	return nil
}

func (p *PacketInfo) Decode(pkt gopacket.Packet, version uint8) *Header { //, key string, IfIndex int32) {
	/*
		gblInfo := p.vrrpGblInfo[key]
		gblInfo.StateNameLock.Lock()
		if gblInfo.StateName == VRRP_INITIALIZE_STATE {
			gblInfo.StateNameLock.Unlock()
			return
		}
		gblInfo.StateNameLock.Unlock()
	*/
	// Check dmac address from the inPacket and if it is same discard the pkt
	ethLayer := pkt.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		svr.logger.Err("Not an eth packet?")
		return
	}
	eth := ethLayer.(*layers.Ethernet)
	// Get Entire IP layer Info
	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		debug.Logger.Err("Not an ip packet?")
		return
	}
	// Get Ip Hdr and start doing basic check according to RFC
	ipHdr := ipLayer.(*layers.IPv4)
	if ipHdr.TTL != VRRP_TTL {
		debug.Logger.Err("ttl should be 255 instead of", ipHdr.TTL, "dropping packet from", ipHdr.SrcIP)
		return
	}
	// Get Payload as checks are succesful
	ipPayload := ipLayer.LayerPayload()
	if ipPayload == nil {
		debug.Logger.Err("No payload for ip packet")
		return
	}
	// Get VRRP header from IP Payload
	hdr := DecodeHeader(ipPayload)
	// Do Basic Vrrp Header Check
	if err := p.ValidateHeader(hdr, ipPayload, key); err != nil {
		debug.Logger.Err(err.Error(), ". Dropping received packet from", ipHdr.SrcIP)
		return
	}

	return hdr
	/*
		// Start FSM for VRRP after all the checks are successful
		p.vrrpFsmCh <- VrrpFsm{
			vrrpHdr: hdr,
			inPkt:   packet,
			key:     key,
		}
	*/
}
