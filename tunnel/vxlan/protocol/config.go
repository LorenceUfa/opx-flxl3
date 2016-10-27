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

// config.go
// Config entry is based on thrift data structures.
package vxlan

import (
	"fmt"
	"net"
	//"strings"
	"errors"
	"vxland"
)

const (
	VxlanCommandCreate = iota + 1
	VxlanCommandDelete
	VxlanCommandUpdate
)

type VxLanConfigChannels struct {
	Vxlancreate               chan VxlanConfig
	Vxlandelete               chan VxlanConfig
	Vxlanupdate               chan VxlanUpdate
	Vtepcreate                chan VtepConfig
	Vtepdelete                chan VtepConfig
	Vtepupdate                chan VtepUpdate
	VxlanAccessPortVlanUpdate chan VxlanAccessPortVlan
	VxlanNextHopUpdate        chan VxlanNextHopIp
	VxlanPortCreate           chan PortConfig
	Vxlanintfinfo             chan VxlanIntfInfo
}

type VxlanIntfInfo struct {
	Command  int
	IntfName string
	IfIndex  int32
	Mac      net.HardwareAddr
	Ip       net.IP
}

type VxlanNextHopIp struct {
	Command   int
	Ip        net.IP
	Intf      int32
	IntfName  string
	NextHopIp net.IP
}

type VxlanAccessPortVlan struct {
	Command  int
	VlanId   uint16
	IntfList []int32
}

type VxlanUpdate struct {
	Oldconfig VxlanConfig
	Newconfig VxlanConfig
	Attr      []string
}

type VtepUpdate struct {
	Oldconfig VtepConfig
	Newconfig VtepConfig
	Attr      []string
}

// bridges attached to the VNI, mapping table to know
// what vlan maps to what VNI used for filtering packets on RX
type VxlanConfig struct {
	VNI          uint32
	UntaggedVlan []uint16
	VlanId       []uint16 // used to tag inner ethernet frame when egressing
	Enable       bool
}

type PortConfig struct {
	Name         string
	HardwareAddr net.HardwareAddr
	Speed        int32
	PortNum      int32
	IfIndex      int32
}

// tunnel endpoint for the VxLAN
type VtepConfig struct {
	Vni                   uint32           `SNAPROUTE: KEY` //VxLAN ID.
	VtepName              string           //VTEP instance name.
	SrcIfName             string           //Source interface ifIndex.
	UDP                   uint16           //vxlan udp port.  Deafult is the iana default udp port
	TTL                   uint16           //TTL of the Vxlan tunnel
	TOS                   uint16           //Type of Service
	MTU                   uint32           //Maximum Transmission Unit
	InnerVlanHandlingMode int32            //The inner vlan tag handling mode.
	Learning              bool             //specifies if unknown source link layer  addresses and IP addresses are entered into the VXLAN  device forwarding database.
	Rsc                   bool             //specifies if route short circuit is turned on.
	L2miss                bool             //specifies if netlink LLADDR miss notifications are generated.
	L3miss                bool             //specifies if netlink IP ADDR miss notifications are generated.
	TunnelSrcIp           net.IP           //Source IP address for the static VxLAN tunnel
	TunnelDstIp           net.IP           //Destination IP address for the static VxLAN tunnel
	VlanId                uint16           //Vlan Id to encapsulate with the vtep tunnel ethernet header
	TunnelSrcMac          net.HardwareAddr //Src Mac assigned to the VTEP within this VxLAN. If an address is not assigned the the local switch address will be used.
	TunnelDstMac          net.HardwareAddr // Optional - may be looked up based on TunnelNextHopIp
	TunnelNextHopIP       net.IP           // NextHopIP is used to find the DMAC for the tunnel within Asicd
	Enable                bool
}

func ConvertInt32ToBool(val int32) bool {
	if val == 0 {
		return false
	}
	return true
}

// VxlanConfigCheck
// Validate the VXLAN provisioning
func VxlanConfigCheck(c *VxlanConfig) error {
	if GetVxlanDBEntry(c.VNI) != nil {
		return errors.New(fmt.Sprintln("Error VxlanInstance Exists vni is not unique", c))
	}
	if len(c.VlanId) > 1 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Only support one VNI->VLAN map", c))
	}
	if len(c.UntaggedVlan) > 1 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Only support one VNI->VLAN map", c))
	}
	if len(c.UntaggedVlan) > 0 && len(c.VlanId) > 0 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Only support one VNI->VLAN map", c))
	}

	// could optimize by making this call once on startup and listen for vlan create
	// updates from asicd but for now this is the simple call (will only affect speed
	// of provisioning or if asicd is not available)
	var vlanList []uint16
	for _, client := range ClientIntf {
		vlanList = client.GetAllVlans()
	}

	unprovvlanlist := make([]uint16, 0)
	for _, vlan := range c.VlanId {
		foundVlan := false

		for _, provvlan := range vlanList {
			if vlan == provvlan {
				foundVlan = true
			}
		}
		if !foundVlan {
			unprovvlanlist = append(unprovvlanlist, vlan)
		}
	}
	if len(unprovvlanlist) > 0 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Prov failed because following vlans were not provisioned %#v", unprovvlanlist))
	}

	return nil
}

func VxlanConfigUpdateCheck(oc *VxlanConfig, nc *VxlanConfig) error {
	if oc.VNI != nc.VNI {
		return errors.New(fmt.Sprintln("Error Unsupported Attribute VNI Update, must delete then create"))
	}
	vxlan := GetVxlanDBEntry(nc.VNI)
	if vxlan == nil {
		return errors.New(fmt.Sprintln("Error Error VxlanInstance Does not exists"))
	}

	if len(nc.VlanId) > 1 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Update Only support one VNI->VLAN map", nc))
	}
	if len(nc.UntaggedVlan) > 1 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Update Only support one VNI->VLAN map", nc))
	}
	if len(nc.UntaggedVlan) > 0 && len(nc.VlanId) > 0 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Update Only support one VNI->VLAN map", nc))
	}

	// could optimize by making this call once on startup and listen for vlan create
	// updates from asicd but for now this is the simple call (will only affect speed
	// of provisioning or if asicd is not available)
	var vlanList []uint16
	for _, client := range ClientIntf {
		vlanList = client.GetAllVlans()
	}

	unprovvlanlist := make([]uint16, 0)
	for _, vlan := range nc.VlanId {
		foundVlan := false
		for _, provvlan := range vlanList {
			if vlan == provvlan {
				foundVlan = true
				break
			}
		}
		if !foundVlan {
			unprovvlanlist = append(unprovvlanlist, vlan)
		}
	}
	if len(unprovvlanlist) > 0 {
		return errors.New(fmt.Sprintln("Error VxlanInstance Prov failed because following vlans were not provisioned %#v", unprovvlanlist))
	}

	return nil
}

// VtepConfigCheck
// Validate the VTEP provisioning
func VtepConfigCheck(c *VtepConfig, create bool) error {
	key := VtepDbKey{
		Name: c.VtepName,
		Vni:  c.Vni,
	}
	if GetVtepDBEntry(&key) != nil && create {
		return errors.New(fmt.Sprintln("Error VtepInstance Exists name is not unique", c))
	}

	if c.TTL > 255 {
		return errors.New("Error VtepInstance TTL must be between 0-255")
	}

	return nil
}

// ConvertVxlanInstanceToVxlanConfig:
// Convert thrift struct to vxlan config and check that db entry exists already
func ConvertVxlanInstanceToVxlanConfig(c *vxland.VxlanInstance, create bool) (*VxlanConfig, error) {

	if !create &&
		GetVxlanDBEntry(uint32(c.Vni)) == nil {
		return nil, errors.New(fmt.Sprintln("Error VxlanInstance does not Exists", c))

	}

	if c.AdminState != "UP" &&
		c.AdminState != "DOWN" {
		return nil, errors.New(fmt.Sprintln("Error VxlanInstance Error unsupported Admin State String must be UP/DOWN"))
	}

	vxlan := &VxlanConfig{

		VNI: uint32(c.Vni),
	}
	for _, untagvlan := range c.UntaggedVlanId {
		vxlan.UntaggedVlan = append(vxlan.UntaggedVlan, uint16(untagvlan))
	}

	for _, vlan := range c.VlanId {
		vxlan.VlanId = append(vxlan.VlanId, uint16(vlan))
	}
	vxlan.Enable = false
	if c.AdminState == "UP" {
		vxlan.Enable = true
	}

	return vxlan, nil
}

func getVtepName(intf string) string {
	vtepName := intf
	//if !strings.Contains("vtep", intf) {
	//	vtepName = "vtep" + intf
	//}
	return vtepName
}

// ConvertVxlanVtepInstanceToVtepConfig:
// Convert thrift struct to vxlan config
func ConvertVxlanVtepInstanceToVtepConfig(c *vxland.VxlanVtepInstance) (*VtepConfig, error) {

	var mac net.HardwareAddr
	var ip net.IP
	var name string
	//var ok bool
	vtepName := getVtepName(c.Intf)
	name = c.IntfRef
	ip = net.ParseIP(c.SrcIp)

	/* TODO need to create a generic way to get an interface name, mac, ip
	if c.SrcIp == "0.0.0.0" && c.IntfRef != "" {
		// need to get the appropriate IntfRef type
		ok, name, mac, ip = snapclient.asicDGetLoopbackInfo()
		if !ok {
			errorstr := "VTEP: Src Tunnel Info not provisioned yet, loopback intf needed"
			logger.Info(errorstr)
			return &VtepConfig{}, errors.New(errorstr)
		}
		fmt.Println("loopback info:", name, mac, ip)
		if c.SrcIp != "0.0.0.0" {
			ip = net.ParseIP(c.SrcIp)
		}
		logger.Info(fmt.Sprintf("Forcing Vtep %s to use Lb %s SrcMac %s Ip %s", vtepName, name, mac, ip))
	}
	*/
	if c.AdminState != "UP" &&
		c.AdminState != "DOWN" {
		return nil, errors.New(fmt.Sprintln("Error VxlanInstance Error unsupported Admin State String must be UP/DOWN"))
	}
	enable := false
	if c.AdminState == "UP" {
		enable = true
	}

	return &VtepConfig{
		Enable:    enable,
		Vni:       uint32(c.Vni),
		VtepName:  vtepName,
		SrcIfName: name,
		UDP:       uint16(c.DstUDP),
		TTL:       uint16(c.TTL),
		TOS:       uint16(c.TOS),
		InnerVlanHandlingMode: c.InnerVlanHandlingMode,
		TunnelSrcIp:           ip,
		TunnelDstIp:           net.ParseIP(c.DstIp),
		VlanId:                uint16(c.VlanId),
		TunnelSrcMac:          mac,
	}, nil
}

func (s *VXLANServer) updateThriftVxLAN(c *VxlanUpdate) {
	// important to note that the attrset starts at index 0 which is the BaseObj
	// which is not the first element on the thrift obj, thus we need to skip
	// this attribute
	updateCfg := false
	for _, objName := range c.Attr {
		if objName == "VlanId" ||
			objName == "UntaggedVlanId" {
			logger.Info(fmt.Sprintf("Updating vlan list tag: %#v,", c.Newconfig.VlanId))
			updateCfg = true
		}
	}

	if updateCfg {
		newvlanlist := make([]uint16, 0)
		delvlanlist := make([]uint16, 0)
		newuntagvlanlist := make([]uint16, 0)
		deluntagvlanlist := make([]uint16, 0)

		vxlan, ok := GetVxlanDB()[c.Newconfig.VNI]
		if ok {
			for idx, nvlan := range c.Newconfig.VlanId {
				foundvlan := false
				for _, ovlan := range c.Oldconfig.VlanId {
					if nvlan == ovlan {
						foundvlan = true
						break
					}
				}
				if !foundvlan {
					newvlanlist = append(newvlanlist, c.Newconfig.VlanId[idx])
				}
			}
			for idx, nvlan := range c.Oldconfig.VlanId {
				foundvlan := false
				for _, ovlan := range c.Newconfig.VlanId {
					if nvlan == ovlan {
						foundvlan = true
						break
					}
				}
				if !foundvlan {
					delvlanlist = append(delvlanlist, c.Oldconfig.VlanId[idx])
				}
			}

			for idx, nvlan := range c.Newconfig.UntaggedVlan {
				foundvlan := false
				for _, ovlan := range c.Oldconfig.UntaggedVlan {
					if nvlan == ovlan {
						foundvlan = true
						break
					}
				}
				if !foundvlan {
					newuntagvlanlist = append(newuntagvlanlist, c.Newconfig.UntaggedVlan[idx])
				}
			}
			for idx, nvlan := range c.Oldconfig.UntaggedVlan {
				foundvlan := false
				for _, ovlan := range c.Newconfig.UntaggedVlan {
					if nvlan == ovlan {
						foundvlan = true
						break
					}
				}
				if !foundvlan {
					deluntagvlanlist = append(deluntagvlanlist, c.Oldconfig.UntaggedVlan[idx])
				}
			}

			// TODO add lock around db as it is used as a filtering
			if len(delvlanlist) > 0 {
				for _, dv := range delvlanlist {
					for idx, vlan := range vxlan.VlanId {
						if dv == vlan {
							vxlan.VlanId = append(vxlan.VlanId[:idx], vxlan.VlanId[idx+1:]...)
							break
						}
					}
				}
			}

			if len(newvlanlist) > 0 {
				for _, nv := range newvlanlist {
					vxlan.VlanId = append(vxlan.VlanId, nv)
				}
			}

			for _, client := range ClientIntf {
				client.UpdateVxlan(vxlan.VNI, newvlanlist, delvlanlist, newuntagvlanlist, deluntagvlanlist)
			}
		}
	}
}

func (s *VXLANServer) updateThriftVtep(c *VtepUpdate) {

	enableobj := false
	disableobj := false
	recreateobj := false
	for _, objName := range c.Attr {
		if objName == "AdminState" {
			if c.Newconfig.Enable {
				enableobj = true
			} else {
				disableobj = true
			}
		} else {
			recreateobj = true
		}
	}

	if disableobj {
		key := &VtepDbKey{
			Name: c.Newconfig.VtepName,
			Vni:  c.Newconfig.Vni,
		}
		vtep := GetVtepDBEntry(key)
		if vtep != nil {
			DeProvisionVtep(vtep, false)
			saveVtepConfigData(&(c.Newconfig))
		}
	} else if enableobj {
		key := &VtepDbKey{
			Name: c.Newconfig.VtepName,
			Vni:  c.Newconfig.Vni,
		}
		vtep := GetVtepDBEntry(key)
		if vtep != nil {
			saveVtepConfigData(&(c.Newconfig))
			ReProvisionVtep(vtep)
		}
	} else if recreateobj {
		DeleteVtep(&(c.Oldconfig))
		CreateVtep(&(c.Newconfig))
	}
}

func (s *VXLANServer) ConfigListener() {

	go func(cc *VxLanConfigChannels) {
		for {
			select {

			case daemonstatus := <-s.DaemonStatusCh:
				if daemonstatus.Name == "asicd" {
					// TODO do something
				} else if daemonstatus.Name == "ribd" {
					// TODO do something
				} else if daemonstatus.Name == "arpd" {
					// TODO do something
				}
			case vxlan := <-cc.Vxlancreate:
				CreateVxLAN(&vxlan)

			case vxlan := <-cc.Vxlandelete:
				DeleteVxLAN(&vxlan)

			case vxlan := <-cc.Vxlanupdate:
				s.updateThriftVxLAN(&vxlan)

			case vtep := <-cc.Vtepcreate:
				CreateVtep(&vtep)

			case vtep := <-cc.Vtepdelete:
				DeleteVtep(&vtep)

			case <-cc.Vtepupdate:
				//s.UpdateThriftVtep(&vtep)

			case <-cc.VxlanAccessPortVlanUpdate:
				// updates from client which are post create of vxlan

			case ipinfo := <-cc.VxlanNextHopUpdate:
				// updates from client which are triggered post create of vtep
				reachable := false
				if ipinfo.Command == VxlanCommandCreate {
					reachable = true
				}
				//ip := net.ParseIP(fmt.Sprintf("%s.%s.%s.%s", uint8(ipinfo.Ip>>24&0xff), uint8(ipinfo.Ip>>16&0xff), uint8(ipinfo.Ip>>8&0xff), uint8(ipinfo.Ip>>0&0xff)))
				HandleNextHopChange(ipinfo.Ip, ipinfo.NextHopIp, ipinfo.Intf, ipinfo.IntfName, reachable)

			case port := <-cc.VxlanPortCreate:
				// store all the valid physical ports
				if _, ok := PortConfigMap[port.IfIndex]; !ok {
					var portcfg = &PortConfig{
						Name:         port.Name,
						HardwareAddr: port.HardwareAddr,
						Speed:        port.Speed,
						PortNum:      port.PortNum,
						IfIndex:      port.IfIndex,
					}
					//logger.Info("Saving Port Config to db", *portcfg)
					PortConfigMap[port.IfIndex] = portcfg
				}
			case intfinfo := <-cc.Vxlanintfinfo:
				for _, vtep := range GetVtepDB() {
					logger.Info(fmt.Sprintln("received intf info", intfinfo, vtep))
					if vtep.SrcIfName == intfinfo.IntfName {

						vtep.VxlanVtepMachineFsm.VxlanVtepEvents <- MachineEvent{
							E:    VxlanVtepEventSrcInterfaceResolved,
							Src:  VxlanVtepMachineModuleStr,
							Data: intfinfo,
						}
					}
				}
			}
		}
	}(s.Configchans)
}
