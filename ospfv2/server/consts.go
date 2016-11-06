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
	"time"
)

const (
	snapshotLen int32         = 65549 //packet capture length
	promiscuous bool          = false //mode
	pcapTimeout time.Duration = 5 * time.Second
)

const (
	OSPF_HELLO_MIN_SIZE  = 20
	OSPF_DBD_MIN_SIZE    = 8
	OSPF_LSA_HEADER_SIZE = 20
	OSPF_LSA_REQ_SIZE    = 12
	OSPF_LSA_ACK_SIZE    = 20
	OSPF_HEADER_SIZE     = 24
	IP_HEADER_MIN_LEN    = 20
	OSPF_PROTO_ID        = 89
	OSPF_VERSION_2       = 2
	OSPF_NO_OF_LSA_FIELD = 4
)
