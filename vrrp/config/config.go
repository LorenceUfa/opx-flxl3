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

package config

const (
	_ = iota
	_ // skipping value 1
	VERSION2
	VERSION3
)

const (
	_ = iota
	CREATE
	UDPATE
	DELETE
)

type PhyPort struct {
	IfIndex   int32
	IntfRef   string
	OperState string
	MacAddr   string
}

type VlanInfo struct {
	Id            int32
	IfIndex       int32
	Name          string
	UntagPortsMap map[int32]bool
	TagPortsMap   map[int32]bool
	OperState     string
}

type IntfCfg struct {
	IntfRef               string
	IfIndex               int32
	VRID                  int32
	Priority              int32
	VirtualIPv4Addr       string
	AdvertisementInterval int32
	PreemptMode           bool
	AcceptMode            bool
	// Information that will be used by server.. as all configs will be passed onto one channel only
	Version   uint8
	Operation uint8
}
