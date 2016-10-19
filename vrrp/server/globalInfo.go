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
	"asicdServices"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	nanomsg "github.com/op/go-nanomsg"
	"l3/vrrp/config"
	"net"
	"sync"
	"time"
	"utils/dbutils"
	"utils/logging"
	"vrrpd"
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

type VrrpPktHeader struct {
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
*/

type VrrpFsm struct {
	key     string
	vrrpHdr *VrrpPktHeader
	inPkt   gopacket.Packet
}

type VrrpClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

type VrrpClientBase struct {
	Address            string
	Transport          thrift.TTransport
	PtrProtocolFactory *thrift.TBinaryProtocolFactory
	IsConnected        bool
}

type VrrpAsicdClient struct {
	VrrpClientBase
	ClientHdl *asicdServices.ASICDServicesClient
}

type VrrpUpdateConfig struct {
	OldConfig vrrpd.VrrpIntf
	NewConfig vrrpd.VrrpIntf
	AttrSet   []bool
}

type VrrpGlobalStateInfo struct {
	AdverRx             uint32 // Total advertisement received
	AdverTx             uint32 // Total advertisement send out
	MasterIp            string // Remote Master Ip Address
	LastAdverRx         string // Last advertisement received
	LastAdverTx         string // Last advertisment send out
	PreviousFsmState    string // previous fsm state
	CurrentFsmState     string // current fsm state
	ReasonForTransition string // why did we transition to current state?
}

type VrrpPktChannelInfo struct {
	pkt     gopacket.Packet
	key     string
	IfIndex int32
}

type VrrpTxChannelInfo struct {
	key      string
	priority uint16 // any value > 255 means ignore it
}

type VrrpServer struct {
	L2Port   map[int32]config.PhyPort
	VlanInfo map[int32]config.VlanInfo
	L3Port   map[int32]IpIntf
	CfgCh    chan *config.IntfCfg
	//	logger                        *logging.Writer
	//	vrrpDbHdl                     *dbutils.DBUtil
	//paramsDir                     string
	//asicdClient                   VrrpAsicdClient
	//asicdSubSocket                *nanomsg.SubSocket
	vrrpGblInfo                   map[string]VrrpGlobalInfo // IfIndex + VRID
	vrrpIntfStateSlice            []string
	vrrpLinuxIfIndex2AsicdIfIndex map[int32]*net.Interface
	vrrpIfIndexIpAddr             map[int32]string
	vrrpVlanId2Name               map[int]string
	VrrpCreateIntfConfigCh        chan vrrpd.VrrpIntf
	VrrpDeleteIntfConfigCh        chan vrrpd.VrrpIntf
	VrrpUpdateIntfConfigCh        chan VrrpUpdateConfig
	vrrpRxPktCh                   chan VrrpPktChannelInfo
	vrrpTxPktCh                   chan VrrpTxChannelInfo
	vrrpFsmCh                     chan VrrpFsm
	vrrpMacConfigAdded            bool
	vrrpSnapshotLen               int32
	vrrpPromiscuous               bool
	vrrpTimeout                   time.Duration
	vrrpPktSend                   chan bool
}

const (
	VRRP_REDDIS_DB_PORT = ":6379"
	VRRP_INTF_DB        = "VrrpIntf"

	// Error Message
	VRRP_INVALID_VRID                   = "VRID is invalid"
	VRRP_CLIENT_CONNECTION_NOT_REQUIRED = "Connection to Client is not required"
	VRRP_SAME_OWNER                     = "Local Router should not be same as the VRRP Ip Address"
	VRRP_MISSING_VRID_CONFIG            = "VRID is not configured on interface"
	VRRP_CHECKSUM_ERR                   = "VRRP checksum failure"
	VRRP_INVALID_PCAP                   = "Invalid Pcap Handler"
	VRRP_VLAN_NOT_CREATED               = "Create Vlan before configuring VRRP"
	VRRP_IPV4_INTF_NOT_CREATED          = "Create IPv4 interface before configuring VRRP"
	VRRP_DATABASE_LOCKED                = "database is locked"

	// VRRP multicast ip address for join
	VRRP_BPF_FILTER = "ip host " + VRRP_GROUP_IP
	VRRP_MAC_MASK   = "ff:ff:ff:ff:ff:ff"
	VRRP_PROTO_ID   = 112

	// Default Size
	VRRP_GLOBAL_INFO_DEFAULT_SIZE         = 50
	VRRP_VLAN_MAPPING_DEFAULT_SIZE        = 5
	VRRP_INTF_STATE_SLICE_DEFAULT_SIZE    = 5
	VRRP_LINUX_INTF_MAPPING_DEFAULT_SIZE  = 5
	VRRP_INTF_IPADDR_MAPPING_DEFAULT_SIZE = 5
	VRRP_RX_BUF_CHANNEL_SIZE              = 100
	VRRP_TX_BUF_CHANNEL_SIZE              = 1
	VRRP_FSM_CHANNEL_SIZE                 = 1
	VRRP_INTF_CONFIG_CH_SIZE              = 1
	VRRP_TOTAL_INTF_CONFIG_ELEMENTS       = 7

	VRRP_MASTER_PRIORITY      = 255
	VRRP_IGNORE_PRIORITY      = 65535
	VRRP_MASTER_DOWN_PRIORITY = 0

	// vrrp default configs
	VRRP_DEFAULT_PRIORITY = 100
	VRRP_IEEE_MAC_ADDR    = "00-00-5E-00-01-"

	// vrrp state names
	VRRP_UNINTIALIZE_STATE = "Un-Initialize"
	VRRP_INITIALIZE_STATE  = "Initialize"
	VRRP_BACKUP_STATE      = "Backup"
	VRRP_MASTER_STATE      = "Master"
)
