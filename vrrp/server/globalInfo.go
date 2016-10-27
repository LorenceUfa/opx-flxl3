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
	"l3/vrrp/packet"
	"net"
	"sync"
	"time"
	"utils/asicdClient"
	"utils/dbutils"
	"utils/dmnBase"
	"utils/logging"
)

type VrrpTxChannelInfo struct {
	key      string
	priority uint16 // any value > 255 means ignore it
}

type KeyInfo struct {
	IntfRef string
	VRID    int32
	Version uint8
}

type IntfState struct {
	State StateInfo
	Key   KeyInfo
}

var vrrpServer *VrrpServer

type VrrpServer struct {
	// All System Related Information
	dmnBase            *dmnBase.FSBaseDmn
	SwitchPlugin       asicdClient.AsicdClientIntf
	GlobalConfig       config.GlobalConfig
	L2Port             map[int32]config.PhyPort
	VlanInfo           map[int32]config.VlanInfo
	V4                 map[int32]*V4Intf
	V6                 map[int32]*V6Intf
	Intf               map[KeyInfo]VrrpInterface // key is struct IntfRef, VRID, Version which is KeyInfo
	V4IntfRefToIfIndex map[string]int32          // key is intfRef and value is ifIndex
	V6IntfRefToIfIndex map[string]int32          // key is intfRef and valud if ifIndex
	CfgCh              chan *config.IntfCfg      // Starting from hereAll Channels Used during Events
	StateCh            chan *IntfState
	GblCfgCh           chan *config.GlobalConfig
}

const (

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
	VRRP_BPF_FILTER      = "ip host " + packet.VRRP_GROUP_IP
	VRRP_MAC_MASK        = "ff:ff:ff:ff:ff:ff"
	VRRP_PROTO_ID        = 112
	VRRP_SNAPSHOT_LEN    = 1024
	VRRP_PROMISCOUS_MODE = false
	VRRP_TIMEOUT         = 1 // in seconds

	// Default Size
	VRRP_GLOBAL_INFO_DEFAULT_SIZE         = 50
	VRRP_VLAN_MAPPING_DEFAULT_SIZE        = 5
	VRRP_INTF_STATE_SLICE_DEFAULT_SIZE    = 5
	VRRP_LINUX_INTF_MAPPING_DEFAULT_SIZE  = 5
	VRRP_INTF_IPADDR_MAPPING_DEFAULT_SIZE = 5
	VRRP_RX_BUF_CHANNEL_SIZE              = 5
	VRRP_TX_BUF_CHANNEL_SIZE              = 1
	VRRP_FSM_CHANNEL_SIZE                 = 1
	VRRP_INTF_CONFIG_CH_SIZE              = 1
	VRRP_TOTAL_INTF_CONFIG_ELEMENTS       = 7

	VRRP_IEEE_MAC_ADDR = "00-00-5E-00-01-"
)
