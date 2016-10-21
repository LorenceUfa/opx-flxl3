// vxlandb.go
package vxlan

import (
	"fmt"
	"net"
)

// vni -> db entry
var vxlanDB map[uint32]*VxlanDbEntry
var vxlanDbList []*VxlanDbEntry

// vxlanDbEntry
// Struct to store the data associated with vxlan
type VxlanDbEntry struct {
	// VNI associated with the vxlan domain
	VNI uint32
	// VlanId associated with the Access endpoints
	VlanId []uint16 // used to tag inner ethernet frame when egressing
	// Multicast IP group (NOT SUPPORTED)
	Group net.IP
	// Shortcut to apply MTU to each VTEP
	MTU uint32
	// Admin State
	Enable bool
}

// NewVxlanDbEntry:
// Create a new vxlan db entry
func NewVxlanDbEntry(c *VxlanConfig) *VxlanDbEntry {
	return &VxlanDbEntry{
		VNI:    c.VNI,
		VlanId: c.VlanId,
		Enable: c.Enable,
	}
}

func GetVxlanDBEntry(vni uint32) *VxlanDbEntry {
	if vxlan, ok := vxlanDB[vni]; ok {
		return vxlan
	}
	return nil
}

func GetVxlanDbListEntry(idx int32, vxlan **VxlanDbEntry) bool {
	if int(idx) < len(vxlanDbList) {
		*vxlan = vxlanDbList[idx]
		return true
	}
	return false
}

// GetVxlanDB:
// returns the vxlan db
func GetVxlanDB() map[uint32]*VxlanDbEntry {
	return vxlanDB
}

// saveVxLanConfigData:
// function saves off the configuration data and saves off the vlan to vni mapping
func saveVxLanConfigData(c *VxlanConfig) {
	if _, ok := vxlanDB[c.VNI]; !ok {
		vxlan := NewVxlanDbEntry(c)
		vxlanDB[c.VNI] = vxlan
	}
}

// DeleteVxLAN:
// Configuration interface for creating the vlxlan instance
func CreateVxLAN(c *VxlanConfig) {
	saveVxLanConfigData(c)

	if VxlanGlobalStateGet() == VXLAN_GLOBAL_ENABLE &&
		c.Enable {
		for _, client := range ClientIntf {
			// create vxlan resources in hw
			client.CreateVxlan(c)
		}

		// lets find all the vteps which are in VtepStatusConfigPending state
		// and initiate a hwConfig
		for _, vtep := range GetVtepDB() {
			if vtep.VxlanVtepMachineFsm.Machine.Curr.CurrentState() == VxlanVtepStateDetached {
				// restart the state machine
				vtep.VxlanVtepMachineFsm.VxlanVtepEvents <- MachineEvent{
					E:   VxlanVtepEventBegin,
					Src: VxlanVtepMachineModuleStr,
				}
			}
		}
		logger.Info(fmt.Sprintln("CreateVxLAN", c))
	}
}

// DeleteVxLAN:
// Configuration interface for deleting the vlxlan instance
func DeleteVxLAN(c *VxlanConfig) {

	if (VxlanGlobalStateGet() == VXLAN_GLOBAL_ENABLE ||
		VxlanGlobalStateGet() == VXLAN_GLOBAL_DISABLE_PENDING) &&
		c.Enable {
		// delete vxlan resources in hw
		for _, client := range ClientIntf {
			client.DeleteVxlan(c)
		}
		logger.Info(fmt.Sprintln("DeleteVxLAN", c.VNI))
	}
	if VxlanGlobalStateGet() == VXLAN_GLOBAL_ENABLE {
		for idx, vni := range vxlanDbList {
			if vni.VNI == c.VNI {
				vxlanDbList = append(vxlanDbList[:idx], vxlanDbList[idx+1:]...)
				break
			}
		}
		delete(vxlanDB, c.VNI)
	}
}
