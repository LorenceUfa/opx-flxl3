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

// lahandler
package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	vxlan "l3/tunnel/vxlan/protocol"
	"models/objects"
	"reflect"
	"utils/dbutils"
	"utils/logging"
	"vxland"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/garyburd/redigo/redis"
)

const DBName string = "UsrConfDb.db"

type ClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

type VXLANDServiceHandler struct {
	server *vxlan.VXLANServer
	logger *logging.Writer
}

// look up the various other daemons based on c string
func GetClientPort(paramsFile string, c string) int {
	var clientsList []ClientJson

	bytes, err := ioutil.ReadFile(paramsFile)
	if err != nil {
		return 0
	}

	err = json.Unmarshal(bytes, &clientsList)
	if err != nil {
		return 0
	}

	for _, client := range clientsList {
		if client.Name == c {
			return client.Port
		}
	}
	return 0
}

func NewVXLANDServiceHandler(server *vxlan.VXLANServer, logger *logging.Writer) *VXLANDServiceHandler {
	//lacp.LacpStartTime = time.Now()
	// link up/down events for now
	//startEvtHandler()
	handler := &VXLANDServiceHandler{
		server: server,
		logger: logger,
	}

	prevState := vxlan.VxlanGlobalStateGet()
	// lets read the current config and re-play the config
	handler.ReadConfigFromDB(prevState)

	return handler
}

func (v *VXLANDServiceHandler) StartThriftServer() {

	var transport thrift.TServerTransport
	var err error

	fileName := v.server.Paramspath + "clients.json"
	port := GetClientPort(fileName, "vxland")
	if port != 0 {
		addr := fmt.Sprintf("localhost:%d", port)
		transport, err = thrift.NewTServerSocket(addr)
		if err != nil {
			panic(fmt.Sprintf("Failed to create Socket with:", addr))
		}

		processor := vxland.NewVXLANDServicesProcessor(v)
		transportFactory := thrift.NewTBufferedTransportFactory(8192)
		protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
		thriftserver := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

		err = thriftserver.Serve()
		panic(err)
	}
	panic(errors.New("Unable to find vxland port"))
}

func (v *VXLANDServiceHandler) CreateVxlanGlobal(config *vxland.VxlanGlobal) (rv bool, err error) {
	rv = true
	v.logger.Info(fmt.Sprintf("CreateVxlanGlobal (server): %s", config.AdminState))

	prevState := vxlan.VxlanGlobalStateGet()

	if config.AdminState == "UP" {
		vxlan.VxlanGlobalStateSet(vxlan.VXLAN_GLOBAL_ENABLE)
	} else if config.AdminState == "DOWN" {
		vxlan.VxlanGlobalStateSet(vxlan.VXLAN_GLOBAL_DISABLE)
	} else {
		return rv, errors.New(fmt.Sprintln("Error VxlanGlobal unknown Admin State setting", config.AdminState))
	}
	v.ReadConfigFromDB(prevState)
	return rv, err
}

func (v *VXLANDServiceHandler) DeleteVxlanGlobal(config *vxland.VxlanGlobal) (bool, error) {
	return true, nil
}

func (v *VXLANDServiceHandler) UpdateVxlanGlobal(origconfig *vxland.VxlanGlobal, updateconfig *vxland.VxlanGlobal, attrset []bool, op []*vxland.PatchOpInfo) (rv bool, err error) {
	v.logger.Info(fmt.Sprintf("UpdateVxlanGlobal (server): %s", updateconfig.AdminState))
	rv = true
	prevState := vxlan.VxlanGlobalStateGet()

	if updateconfig.AdminState == "UP" {
		vxlan.VxlanGlobalStateSet(vxlan.VXLAN_GLOBAL_ENABLE)
	} else if updateconfig.AdminState == "DOWN" {
		vxlan.VxlanGlobalStateSet(vxlan.VXLAN_GLOBAL_DISABLE_PENDING)
	} else {
		return rv, errors.New(fmt.Sprintln("Error Update VxlanGlobal unknown Admin State setting", updateconfig.AdminState))
	}
	if prevState != vxlan.VxlanGlobalStateGet() {
		v.ReadConfigFromDB(prevState)
	}
	return rv, err
}

func (v *VXLANDServiceHandler) GetBulkVxlanGlobalState(fromIndex vxland.Int, count vxland.Int) (obj *vxland.VxlanGlobalStateGetInfo, err error) {
	var returnVxlanGlobalStates []*vxland.VxlanGlobalState
	var returnVxlanGlobalStateGetInfo vxland.VxlanGlobalStateGetInfo
	toIndex := fromIndex
	obj = &returnVxlanGlobalStateGetInfo

	nextVxlanGlobalState, gserr := v.GetVxlanGlobalState("default")
	if gserr == nil {
		if len(returnVxlanGlobalStates) == 0 {
			returnVxlanGlobalStates = make([]*vxland.VxlanGlobalState, 0)
		}
		returnVxlanGlobalStates = append(returnVxlanGlobalStates, nextVxlanGlobalState)
	}
	obj.VxlanGlobalStateList = returnVxlanGlobalStates
	obj.StartIdx = fromIndex
	obj.EndIdx = toIndex + 1
	obj.More = false
	obj.Count = 1
	return obj, err
}
func (v *VXLANDServiceHandler) GetVxlanGlobalState(Vrf string) (*vxland.VxlanGlobalState, error) {

	vg := &vxland.VxlanGlobalState{}

	state := vxlan.VxlanGlobalStateGet()
	if state == vxlan.VXLAN_GLOBAL_ENABLE {
		vg.OperState = "UP"
	} else {
		vg.OperState = "DOWN"
	}
	vg.RxInvalidVtepCnt = 0
	vg.NumVteps = int64(len(vxlan.GetVtepDB()))
	return vg, nil
}

func (v *VXLANDServiceHandler) CreateVxlanInstance(config *vxland.VxlanInstance) (bool, error) {
	v.logger.Info(fmt.Sprintf("CreateVxlanConfigInstance %#v", config))

	c, err := vxlan.ConvertVxlanInstanceToVxlanConfig(config, true)
	if err == nil {
		err = vxlan.VxlanConfigCheck(c)
		if err == nil {
			v.server.Configchans.Vxlancreate <- *c
			return true, nil
		}
	}
	return false, err
}

func (v *VXLANDServiceHandler) DeleteVxlanInstance(config *vxland.VxlanInstance) (bool, error) {
	v.logger.Info(fmt.Sprintf("DeleteVxlanConfigInstance %#v", config))
	c, err := vxlan.ConvertVxlanInstanceToVxlanConfig(config, false)
	if err == nil {
		v.server.Configchans.Vxlandelete <- *c
		return true, nil
	}
	return false, err
}

func (v *VXLANDServiceHandler) UpdateVxlanInstance(origconfig *vxland.VxlanInstance, newconfig *vxland.VxlanInstance, attrset []bool, op []*vxland.PatchOpInfo) (bool, error) {
	v.logger.Info(fmt.Sprintf("UpdateVxlanConfigInstance orig[%#v] new[%#v] attrset[%#v]", origconfig, newconfig, attrset))
	oc, _ := vxlan.ConvertVxlanInstanceToVxlanConfig(origconfig, false)
	nc, err := vxlan.ConvertVxlanInstanceToVxlanConfig(newconfig, false)
	if err == nil {
		err = vxlan.VxlanConfigUpdateCheck(oc, nc)
		if err == nil {

			strattr := make([]string, 0)
			objTyp := reflect.TypeOf(*origconfig)

			// important to note that the attrset starts at index 0 which is the BaseObj
			// which is not the first element on the thrift obj, thus we need to skip
			// this attribute
			for i := 0; i < objTyp.NumField(); i++ {
				objName := objTyp.Field(i).Name
				if attrset[i] {
					strattr = append(strattr, objName)
				}
			}
			update := vxlan.VxlanUpdate{
				Oldconfig: *oc,
				Newconfig: *nc,
				Attr:      strattr,
			}
			v.server.Configchans.Vxlanupdate <- update
			return true, nil
		}
	}
	return false, err
}

func (v *VXLANDServiceHandler) CreateVxlanVtepInstance(config *vxland.VxlanVtepInstance) (bool, error) {
	v.logger.Info(fmt.Sprintf("CreateVxlanVtepInstance %#v", config))
	c, err := vxlan.ConvertVxlanVtepInstanceToVtepConfig(config)
	if err == nil {
		err = vxlan.VtepConfigCheck(c)
		if err == nil {
			v.server.Configchans.Vtepcreate <- *c
			return true, err
		}
	}
	return false, err
}

func (v *VXLANDServiceHandler) DeleteVxlanVtepInstance(config *vxland.VxlanVtepInstance) (bool, error) {
	v.logger.Info(fmt.Sprintf("DeleteVxlanVtepInstance %#v", config))
	c, err := vxlan.ConvertVxlanVtepInstanceToVtepConfig(config)
	if err == nil {
		v.server.Configchans.Vtepdelete <- *c
		return true, nil
	}
	return false, err
}

func (v *VXLANDServiceHandler) UpdateVxlanVtepInstance(origconfig *vxland.VxlanVtepInstance, newconfig *vxland.VxlanVtepInstance, attrset []bool, op []*vxland.PatchOpInfo) (bool, error) {
	v.logger.Info(fmt.Sprintf("UpdateVxlanVtepInstances orig[%#v] new[%#v]", origconfig, newconfig))
	oc, _ := vxlan.ConvertVxlanVtepInstanceToVtepConfig(origconfig)
	nc, err := vxlan.ConvertVxlanVtepInstanceToVtepConfig(newconfig)
	if err == nil {
		err = vxlan.VtepConfigCheck(nc)
		if err == nil {
			strattr := make([]string, 0)
			objTyp := reflect.TypeOf(*origconfig)

			// important to note that the attrset starts at index 0 which is the BaseObj
			// which is not the first element on the thrift obj, thus we need to skip
			// this attribute
			for i := 0; i < objTyp.NumField(); i++ {
				objName := objTyp.Field(i).Name
				if attrset[i] {
					strattr = append(strattr, objName)
				}
			}

			update := vxlan.VtepUpdate{
				Oldconfig: *oc,
				Newconfig: *nc,
				Attr:      strattr,
			}
			v.server.Configchans.Vtepupdate <- update
			return true, nil
		}
	}

	return false, err
}

func (v *VXLANDServiceHandler) GetVxlanInstanceState(vni int32) (*vxland.VxlanInstanceState, error) {
	vis := &vxland.VxlanInstanceState{}
	if v, ok := vxlan.GetVxlanDB()[uint32(vni)]; ok {
		vis.Vni = int32(vni)
		if v.Enable {
			vis.OperState = "UP"
		} else {
			vis.OperState = "DOWN"
		}
		for _, vlan := range v.VlanId {
			vis.VlanId = append(vis.VlanId, int16(vlan))
		}

	} else {
		return nil, errors.New(fmt.Sprintf("Error could not find vni instance %d", vni))
	}
	return vis, nil
}

func (la *VXLANDServiceHandler) GetBulkVxlanInstanceState(fromIndex vxland.Int, count vxland.Int) (obj *vxland.VxlanInstanceStateGetInfo, err error) {

	var vxlanStateList []vxland.VxlanInstanceState = make([]vxland.VxlanInstanceState, count)
	var nextVxlanState *vxland.VxlanInstanceState
	var returnVxlanStates []*vxland.VxlanInstanceState
	var returnVxlanStateGetInfo vxland.VxlanInstanceStateGetInfo
	validCount := vxland.Int(0)
	toIndex := fromIndex
	moreRoutes := false
	obj = &returnVxlanStateGetInfo

	var v *vxlan.VxlanDbEntry
	currIndex := vxland.Int(0)

	for currIndex = vxland.Int(0); validCount != count && vxlan.GetVxlanDbListEntry(int32(currIndex), &v); currIndex++ {

		if currIndex < fromIndex {
			continue
		} else {

			nextVxlanState = &vxlanStateList[validCount]
			if v.Enable {
				nextVxlanState.OperState = "UP"
			} else {
				nextVxlanState.OperState = "DOWN"
			}
			nextVxlanState.Vni = int32(v.VNI)
			for _, vlan := range v.VlanId {
				nextVxlanState.VlanId = append(nextVxlanState.VlanId, int16(vlan))
			}
			if len(returnVxlanStates) == 0 {
				returnVxlanStates = make([]*vxland.VxlanInstanceState, 0)
			}
			returnVxlanStates = append(returnVxlanStates, nextVxlanState)
			validCount++
			toIndex++
		}
	}
	// lets try and get the next agg if one exists then there are more routes
	if v != nil {
		moreRoutes = vxlan.GetVxlanDbListEntry(int32(currIndex), &v)
	}
	obj.VxlanInstanceStateList = returnVxlanStates
	obj.StartIdx = fromIndex
	obj.EndIdx = toIndex + 1
	obj.More = moreRoutes
	obj.Count = validCount

	return obj, nil
}

func (v *VXLANDServiceHandler) GetVxlanVtepInstanceState(intf string) (*vxland.VxlanVtepInstanceState, error) {
	vis := &vxland.VxlanVtepInstanceState{}
	key := vxlan.VtepDbKey{
		Name: intf,
	}
	if v, ok := vxlan.GetVtepDB()[key]; ok {
		if v.Enable {
			vis.OperState = "UP"
		} else {
			vis.OperState = "DOWN"
		}
		vis.Intf = v.VtepName
		vis.IntfRef = v.SrcIfName
		vis.IfIndex = v.VtepIfIndex
		vis.Vni = int32(v.Vni)
		vis.DstUDP = int16(v.UDP)
		vis.TTL = int16(v.TTL)
		vis.TOS = int16(v.TOS)
		if v.FilterUnknownCustVlan {
			vis.InnerVlanHandlingMode = 0
		} else {
			vis.InnerVlanHandlingMode = 1
		}
		vis.DstIp = v.DstIp.String()
		vis.SrcIp = v.SrcIp.String()
		vis.VlanId = int16(v.VlanId)
		vis.Mtu = int32(v.MTU)
		vis.RxPkts = int64(v.GetRxStats())
		vis.TxPkts = int64(v.GetTxStats())
		//vis.RxFwdPkts             uint64 `DESCRIPTION: Rx Forwaded Packets`
		//vis.RxDropPkts            uint64 `DESCRIPTION: Rx Dropped Packets`
		//vis.RxUnknownVni          uint64 `DESCRIPTION: Rx Unknown Vni in frame`
		vis.VtepFsmState = vxlan.VxlanVtepStateStrMap[v.VxlanVtepMachineFsm.Machine.Curr.CurrentState()]
		vis.VtepFsmPrevState = vxlan.VxlanVtepStateStrMap[v.VxlanVtepMachineFsm.Machine.Curr.PreviousState()]

	} else {
		return nil, errors.New(fmt.Sprintf("Error could not find vni instance %s", intf))
	}
	return vis, nil
}

func (la *VXLANDServiceHandler) GetBulkVxlanVtepInstanceState(fromIndex vxland.Int, count vxland.Int) (obj *vxland.VxlanVtepInstanceStateGetInfo, err error) {

	var vxlanVtepStateList []vxland.VxlanVtepInstanceState = make([]vxland.VxlanVtepInstanceState, count)
	var nextVxlanVtepState *vxland.VxlanVtepInstanceState
	var returnVxlanVtepStates []*vxland.VxlanVtepInstanceState
	var returnVxlanVtepStateGetInfo vxland.VxlanVtepInstanceStateGetInfo
	validCount := vxland.Int(0)
	toIndex := fromIndex
	moreRoutes := false
	obj = &returnVxlanVtepStateGetInfo

	var v *vxlan.VtepDbEntry
	currIndex := vxland.Int(0)
	for currIndex = vxland.Int(0); validCount != count && vxlan.GetVtepDbListEntry(int32(currIndex), &v); currIndex++ {

		if currIndex < fromIndex {
			continue
		} else {

			nextVxlanVtepState = &vxlanVtepStateList[validCount]
			if v.Enable {
				nextVxlanVtepState.OperState = "UP"
			} else {
				nextVxlanVtepState.OperState = "DOWN"
			}
			nextVxlanVtepState.Intf = v.VtepName
			nextVxlanVtepState.IntfRef = v.SrcIfName
			nextVxlanVtepState.IfIndex = v.VtepIfIndex
			nextVxlanVtepState.Vni = int32(v.Vni)
			nextVxlanVtepState.DstUDP = int16(v.UDP)
			nextVxlanVtepState.TTL = int16(v.TTL)
			nextVxlanVtepState.TOS = int16(v.TOS)
			if v.FilterUnknownCustVlan {
				nextVxlanVtepState.InnerVlanHandlingMode = 0
			} else {
				nextVxlanVtepState.InnerVlanHandlingMode = 1
			}
			nextVxlanVtepState.DstIp = v.DstIp.String()
			nextVxlanVtepState.SrcIp = v.SrcIp.String()
			nextVxlanVtepState.VlanId = int16(v.VlanId)
			nextVxlanVtepState.Mtu = int32(v.MTU)
			nextVxlanVtepState.RxPkts = int64(v.GetRxStats())
			nextVxlanVtepState.TxPkts = int64(v.GetTxStats())
			//nextVxlanVtepState.RxFwdPkts             uint64 `DESCRIPTION: Rx Forwaded Packets`
			//nextVxlanVtepState.RxDropPkts            uint64 `DESCRIPTION: Rx Dropped Packets`
			//nextVxlanVtepState.RxUnknownVni          uint64 `DESCRIPTION: Rx Unknown Vni in frame`
			nextVxlanVtepState.VtepFsmState = vxlan.VxlanVtepStateStrMap[v.VxlanVtepMachineFsm.Machine.Curr.CurrentState()]
			nextVxlanVtepState.VtepFsmPrevState = vxlan.VxlanVtepStateStrMap[v.VxlanVtepMachineFsm.Machine.Curr.PreviousState()]

			if len(returnVxlanVtepStates) == 0 {
				returnVxlanVtepStates = make([]*vxland.VxlanVtepInstanceState, 0)
			}
			returnVxlanVtepStates = append(returnVxlanVtepStates, nextVxlanVtepState)
			validCount++
			toIndex++
		}
	}
	// lets try and get the next agg if one exists then there are more routes
	if v != nil {
		moreRoutes = vxlan.GetVtepDbListEntry(int32(currIndex), &v)
	}
	obj.VxlanVtepInstanceStateList = returnVxlanVtepStates
	obj.StartIdx = fromIndex
	obj.EndIdx = toIndex + 1
	obj.More = moreRoutes
	obj.Count = validCount

	return obj, nil
}

func (v *VXLANDServiceHandler) HandleDbReadVxlanGlobal(dbHdl redis.Conn) error {
	if dbHdl != nil {
		var dbObj objects.VxlanGlobal
		objList, err := dbObj.GetAllObjFromDb(dbHdl)
		if err != nil {
			v.logger.Warning("DB Query failed when retrieving VxlanInstance objects")
			return err
		}
		for idx := 0; idx < len(objList); idx++ {
			obj := vxland.NewVxlanGlobal()
			dbObject := objList[idx].(objects.VxlanGlobal)
			objects.ConvertvxlandVxlanGlobalObjToThrift(&dbObject, obj)
			v.CreateVxlanGlobal(obj)
		}
	}
	return nil
}

func (v *VXLANDServiceHandler) HandleDbReadVxlanInstance(dbHdl redis.Conn, disable bool) error {
	if dbHdl != nil {
		var dbObj objects.VxlanInstance
		objList, err := dbObj.GetAllObjFromDb(dbHdl)
		if err != nil {
			v.logger.Warning("DB Query failed when retrieving VxlanInstance objects")
			return err
		}
		for idx := 0; idx < len(objList); idx++ {
			obj := vxland.NewVxlanInstance()
			dbObject := objList[idx].(objects.VxlanInstance)
			objects.ConvertvxlandVxlanInstanceObjToThrift(&dbObject, obj)

			if disable {
				_, err = v.DeleteVxlanInstance(obj)
			} else {
				_, err = v.CreateVxlanInstance(obj)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VXLANDServiceHandler) HandleDbReadVxlanVtepInstance(dbHdl redis.Conn, disable bool) error {
	if dbHdl != nil {
		var dbObj objects.VxlanVtepInstance
		objList, err := dbObj.GetAllObjFromDb(dbHdl)
		if err != nil {
			v.logger.Warning("DB Query failed when retrieving VxlanVtepInstance objects")
			return err
		}
		for idx := 0; idx < len(objList); idx++ {
			obj := vxland.NewVxlanVtepInstance()
			dbObject := objList[idx].(objects.VxlanVtepInstance)
			objects.ConvertvxlandVxlanVtepInstanceObjToThrift(&dbObject, obj)
			if disable {
				_, err = v.DeleteVxlanVtepInstance(obj)
			} else {
				_, err = v.CreateVxlanVtepInstance(obj)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VXLANDServiceHandler) ReadConfigFromDB(prevState int) error {

	dbHdl := dbutils.NewDBUtil(v.logger)
	err := dbHdl.Connect()
	if err != nil {
		v.logger.Err("Unable to connect to db")
		return err
	}
	defer dbHdl.Disconnect()

	if prevState == vxlan.VXLAN_GLOBAL_INIT {

		if err := v.HandleDbReadVxlanGlobal(dbHdl); err != nil {
			fmt.Println("Error getting All LacpGlobal objects")
			return err
		}
	}
	currState := vxlan.VxlanGlobalStateGet()

	v.logger.Info(fmt.Sprintf("Global State prev %d curr %d", prevState, currState))

	if currState == vxlan.VXLAN_GLOBAL_DISABLE_PENDING ||
		prevState == vxlan.VXLAN_GLOBAL_ENABLE {

		// lets delete the Aggregator first
		if err := v.HandleDbReadVxlanVtepInstance(dbHdl, true); err != nil {
			v.logger.Err("Error getting All VxlanVtep objects")
			return err
		}

		if err := v.HandleDbReadVxlanInstance(dbHdl, true); err != nil {
			v.logger.Err("Error getting All Vxlan objects")
			return err
		}
	} else if prevState != currState {

		if err := v.HandleDbReadVxlanInstance(dbHdl, false); err != nil {
			fmt.Println("Error getting All Vxlan objects")
			return err
		}

		if err := v.HandleDbReadVxlanVtepInstance(dbHdl, false); err != nil {
			fmt.Println("Error getting All VxlanVtep objects")
			return err
		}
	}

	return nil
}
