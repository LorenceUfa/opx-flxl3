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
	"asicd/asicdCommonDefs"
	"asicdServices"
	//	"database/sql"
	"fmt"
	"github.com/op/go-nanomsg"
	"l3/rib/ribdCommonDefs"
	"net"
	"os"
	"os/signal"
	"ribd"
	"ribdInt"
	"syscall"
	"utils/dbutils"
	"utils/logging"
	"utils/patriciaDB"
	"utils/policy"
	"utils/policy/policyCommonDefs"
)

type RIBdServerConfig struct {
	OrigConfigObject          interface{}
	NewConfigObject           interface{}
	OrigBulkRouteConfigObject []*ribdInt.IPv4RouteConfig
	Bulk                      bool
	BulkEnd                   bool
	AttrSet                   []bool
	Op                        string //"add"/"del"/"update/get"
	PatchOp                   []*ribd.PatchOpInfo
}

/*type PatchUpdateRouteInfo struct {
	OrigRoute *ribd.IPv4Route
	NewRoute  *ribd.IPv4Route
	Op        []*ribd.PatchOpInfo
}*/
type RIBDServer struct {
	Logger                 *logging.Writer
	PolicyEngineDB         *policy.PolicyEngineDB
	GlobalPolicyEngineDB   *policy.PolicyEngineDB
	TrackReachabilityCh    chan TrackReachabilityInfo
	RouteConfCh            chan RIBdServerConfig
	AsicdRouteCh           chan RIBdServerConfig
	ArpdRouteCh            chan RIBdServerConfig
	NotificationChannel    chan NotificationMsg
	NextHopInfoMap         map[NextHopInfoKey]NextHopInfo
	PolicyConditionConfCh  chan RIBdServerConfig
	PolicyActionConfCh     chan RIBdServerConfig
	PolicyStmtConfCh       chan RIBdServerConfig
	PolicyDefinitionConfCh chan RIBdServerConfig
	PolicyApplyCh          chan ApplyPolicyInfo
	PolicyUpdateApplyCh    chan ApplyPolicyInfo
	DBRouteCh              chan RIBdServerConfig
	AcceptConfig           bool
	ServerUpCh             chan bool
	DBReadDone             chan bool
	DbHdl                  *dbutils.DBUtil
	Clients                map[string]ClientIf
	//RouteInstallCh                 chan RouteParams
}

const (
	PROTOCOL_NONE      = -1
	PROTOCOL_CONNECTED = 0
	PROTOCOL_STATIC    = 1
	PROTOCOL_OSPF      = 2
	PROTOCOL_BGP       = 3
	PROTOCOL_LAST      = 4
)

const (
	add = iota
	del
	delAll
	invalidate
)
const (
	Invalid   = -1
	FIBOnly   = 0
	FIBAndRIB = 1
	RIBOnly   = 2
)
const (
	SUB_ASICD = 0
)

type localDB struct {
	prefix     patriciaDB.Prefix
	isValid    bool
	precedence int
	nextHopIp  string
}
type IntfEntry struct {
	name string
}

var count int
var ConnectedRoutes []*ribdInt.Routes
var logger *logging.Writer
var AsicdSub *nanomsg.SubSocket
var RouteServiceHandler *RIBDServer
var IntfIdNameMap map[int32]IntfEntry
var IfNameToIfIndex map[string]int32
var GlobalPolicyEngineDB *policy.PolicyEngineDB
var PolicyEngineDB *policy.PolicyEngineDB
var PARAMSDIR string
var v4rtCount int
var v4routeCreatedTimeMap map[int]string
var v6rtCount int
var v6routeCreatedTimeMap map[int]string

var dbReqCount = 0
var dbReqCountLimit = 1
var dbReqCheckCount = 0
var dbReqCheckCountLimit = 5

/*
   Handle Interface down event
*/
func (ribdServiceHandler *RIBDServer) ProcessL3IntfDownEvent(ipAddr string) {
	logger.Debug("processL3IntfDownEvent")
	var ipMask net.IP
	ip, ipNet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return
	}
	ipMask = make(net.IP, 4)
	copy(ipMask, ipNet.Mask)
	ipAddrStr := ip.String()
	ipMaskStr := net.IP(ipMask).String()
	logger.Info(fmt.Sprintln(" processL3IntfDownEvent for  ipaddr ", ipAddrStr, " mask ", ipMaskStr))
	for i := 0; i < len(ConnectedRoutes); i++ {
		if ConnectedRoutes[i].Ipaddr == ipAddrStr && ConnectedRoutes[i].Mask == ipMaskStr {
			logger.Info(fmt.Sprintln("Delete this route with destAddress = ", ConnectedRoutes[i].Ipaddr, " nwMask = ", ConnectedRoutes[i].Mask))
			deleteIPRoute(ConnectedRoutes[i].Ipaddr, ConnectedRoutes[i].Mask, "CONNECTED", ConnectedRoutes[i].NextHopIp, FIBOnly, ribdCommonDefs.RoutePolicyStateChangeNoChange)
		}
	}
}

/*
   Handle Interface up event
*/
func (ribdServiceHandler *RIBDServer) ProcessL3IntfUpEvent(ipAddr string) {
	logger.Debug("processL3IntfUpEvent")
	var ipMask net.IP
	ip, ipNet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return
	}
	ipMask = make(net.IP, 4)
	copy(ipMask, ipNet.Mask)
	ipAddrStr := ip.String()
	ipMaskStr := net.IP(ipMask).String()
	logger.Info(" processL3IntfUpEvent for  ipaddr ", ipAddrStr, " mask ", ipMaskStr)
	for i := 0; i < len(ConnectedRoutes); i++ {
		logger.Info("Current state of this connected route is ", ConnectedRoutes[i].IsValid)
		if ConnectedRoutes[i].Ipaddr == ipAddrStr && ConnectedRoutes[i].Mask == ipMaskStr && ConnectedRoutes[i].IsValid == false {
			logger.Info("Add this route with destAddress = ", ConnectedRoutes[i].Ipaddr, " nwMask = ", ConnectedRoutes[i].Mask)

			ConnectedRoutes[i].IsValid = true
			policyRoute := ribdInt.Routes{Ipaddr: ConnectedRoutes[i].Ipaddr, Mask: ConnectedRoutes[i].Mask, NextHopIp: ConnectedRoutes[i].NextHopIp, IfIndex: ConnectedRoutes[i].IfIndex, Metric: ConnectedRoutes[i].Metric, Prototype: ConnectedRoutes[i].Prototype}
			params := RouteParams{destNetIp: ConnectedRoutes[i].Ipaddr, networkMask: ConnectedRoutes[i].Mask, nextHopIp: ConnectedRoutes[i].NextHopIp, nextHopIfIndex: ribd.Int(ConnectedRoutes[i].IfIndex), metric: ribd.Int(ConnectedRoutes[i].Metric), routeType: ribd.Int(ConnectedRoutes[i].Prototype), sliceIdx: ribd.Int(ConnectedRoutes[i].SliceIdx), createType: FIBOnly, deleteType: Invalid}
			PolicyEngineFilter(policyRoute, policyCommonDefs.PolicyPath_Import, params)
		}
	}
}

func getLogicalIntfInfo() {
	logger.Debug("Getting Logical Interfaces from asicd")
	var currMarker asicdServices.Int
	var count asicdServices.Int
	count = 100
	for {
		logger.Info("Getting ", count, "GetBulkLogicalIntf objects from currMarker:", currMarker)
		bulkInfo, err := asicdclnt.ClientHdl.GetBulkLogicalIntfState(currMarker, count)
		if err != nil {
			logger.Info("GetBulkLogicalIntfState with err ", err)
			return
		}
		if bulkInfo.Count == 0 {
			logger.Info("0 objects returned from GetBulkLogicalIntfState")
			return
		}
		logger.Info("len(bulkInfo.GetBulkLogicalIntfState)  = ", len(bulkInfo.LogicalIntfStateList), " num objects returned = ", bulkInfo.Count)
		for i := 0; i < int(bulkInfo.Count); i++ {
			ifId := (bulkInfo.LogicalIntfStateList[i].IfIndex)
			logger.Info("logical interface = ", bulkInfo.LogicalIntfStateList[i].Name, "ifId = ", ifId)
			if IntfIdNameMap == nil {
				IntfIdNameMap = make(map[int32]IntfEntry)
			}
			intfEntry := IntfEntry{name: bulkInfo.LogicalIntfStateList[i].Name}
			IntfIdNameMap[ifId] = intfEntry
			if IfNameToIfIndex == nil {
				IfNameToIfIndex = make(map[string]int32)
			}
			IfNameToIfIndex[bulkInfo.LogicalIntfStateList[i].Name] = ifId
		}
		if bulkInfo.More == false {
			logger.Info("more returned as false, so no more get bulks")
			return
		}
		currMarker = asicdServices.Int(bulkInfo.EndIdx)
	}
}
func getVlanInfo() {
	logger.Debug("Getting vlans from asicd")
	var currMarker asicdServices.Int
	var count asicdServices.Int
	count = 100
	for {
		logger.Info("Getting ", count, "GetBulkVlan objects from currMarker:", currMarker)
		bulkInfo, err := asicdclnt.ClientHdl.GetBulkVlanState(currMarker, count)
		if err != nil {
			logger.Info("GetBulkVlan with err ", err)
			return
		}
		if bulkInfo.Count == 0 {
			logger.Info("0 objects returned from GetBulkVlan")
			return
		}
		logger.Info("len(bulkInfo.GetBulkVlan)  = ", len(bulkInfo.VlanStateList), " num objects returned = ", bulkInfo.Count)
		for i := 0; i < int(bulkInfo.Count); i++ {
			ifId := (bulkInfo.VlanStateList[i].IfIndex)
			logger.Info("vlan = ", bulkInfo.VlanStateList[i].VlanId, "ifId = ", ifId)
			if IntfIdNameMap == nil {
				IntfIdNameMap = make(map[int32]IntfEntry)
			}
			intfEntry := IntfEntry{name: bulkInfo.VlanStateList[i].VlanName}
			IntfIdNameMap[ifId] = intfEntry
			if IfNameToIfIndex == nil {
				IfNameToIfIndex = make(map[string]int32)
			}
			IfNameToIfIndex[bulkInfo.VlanStateList[i].VlanName] = ifId
		}
		if bulkInfo.More == false {
			logger.Info("more returned as false, so no more get bulks")
			return
		}
		currMarker = asicdServices.Int(bulkInfo.EndIdx)
	}
}
func getPortInfo() {
	logger.Debug("Getting ports from asicd")
	var currMarker asicdServices.Int
	var count asicdServices.Int
	count = 100
	for {
		logger.Info("Getting ", count, "objects from currMarker:", currMarker)
		bulkInfo, err := asicdclnt.ClientHdl.GetBulkPortState(currMarker, count)
		if err != nil {
			logger.Info("GetBulkPortState with err ", err)
			return
		}
		if bulkInfo.Count == 0 {
			logger.Info("0 objects returned from GetBulkPortState")
			return
		}
		logger.Info("len(bulkInfo.PortStateList)  = ", len(bulkInfo.PortStateList), " num objects returned = ", bulkInfo.Count)
		for i := 0; i < int(bulkInfo.Count); i++ {
			ifId := bulkInfo.PortStateList[i].IfIndex
			logger.Info("ifId = ", ifId)
			if IntfIdNameMap == nil {
				IntfIdNameMap = make(map[int32]IntfEntry)
			}
			intfEntry := IntfEntry{name: bulkInfo.PortStateList[i].Name}
			IntfIdNameMap[ifId] = intfEntry
			if IfNameToIfIndex == nil {
				IfNameToIfIndex = make(map[string]int32)
			}
			IfNameToIfIndex[bulkInfo.PortStateList[i].Name] = ifId
		}
		if bulkInfo.More == false {
			logger.Info("more returned as false, so no more get bulks")
			return
		}
		currMarker = asicdServices.Int(bulkInfo.EndIdx)
	}
}
func getIntfInfo() {
	getPortInfo()
	getVlanInfo()
	getLogicalIntfInfo()
}
func (ribdServiceHandler *RIBDServer) AcceptConfigActions() {
	logger.Info("AcceptConfigActions: Setting AcceptConfig to true")
	RouteServiceHandler.AcceptConfig = true
	getIntfInfo()
	getConnectedRoutes()
	//update dbRouteCh to fetch route data
	ribdServiceHandler.DBRouteCh <- RIBdServerConfig{Op: "fetch"}
	dbRead := <-ribdServiceHandler.DBReadDone
	logger.Info("Received dbread: ")
	if dbRead != true {
		logger.Err("DB read failed")
	}
	go ribdServiceHandler.SetupEventHandler(AsicdSub, asicdCommonDefs.PUB_SOCKET_ADDR, SUB_ASICD)
	logger.Info("All set to signal start the RIBd server")
	ribdServiceHandler.ServerUpCh <- true
}

func (ribdServiceHandler *RIBDServer) InitializeGlobalPolicyDB() *policy.PolicyEngineDB {
	ribdServiceHandler.GlobalPolicyEngineDB = policy.NewPolicyEngineDB(logger)
	ribdServiceHandler.GlobalPolicyEngineDB.SetDefaultImportPolicyActionFunc(defaultImportPolicyEngineActionFunc)
	ribdServiceHandler.GlobalPolicyEngineDB.SetDefaultExportPolicyActionFunc(defaultExportPolicyEngineActionFunc)
	ribdServiceHandler.GlobalPolicyEngineDB.SetIsEntityPresentFunc(DoesRouteExist)
	ribdServiceHandler.GlobalPolicyEngineDB.SetEntityUpdateFunc(UpdateRouteAndPolicyDB)
	ribdServiceHandler.GlobalPolicyEngineDB.SetActionFunc(policyCommonDefs.PolicyActionTypeRouteDisposition, policyEngineRouteDispositionAction)
	ribdServiceHandler.GlobalPolicyEngineDB.SetActionFunc(policyCommonDefs.PolicyActionTypeRouteRedistribute, policyEngineActionRedistribute)
	ribdServiceHandler.GlobalPolicyEngineDB.SetActionFunc(policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise, policyEngineActionNetworkStatementAdvertise)
	ribdServiceHandler.GlobalPolicyEngineDB.SetActionFunc(policyCommonDefs.PoilcyActionTypeSetAdminDistance, policyEngineActionSetAdminDistance)
	ribdServiceHandler.GlobalPolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PolicyActionTypeRouteDisposition, policyEngineUndoRouteDispositionAction)
	ribdServiceHandler.GlobalPolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PolicyActionTypeRouteRedistribute, policyEngineActionUndoRedistribute)
	ribdServiceHandler.GlobalPolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PoilcyActionTypeSetAdminDistance, policyEngineActionUndoSetAdminDistance)
	ribdServiceHandler.GlobalPolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise, policyEngineActionUndoNetworkStatemenAdvertiseAction)
	ribdServiceHandler.GlobalPolicyEngineDB.SetTraverseAndApplyPolicyFunc(policyEngineTraverseAndApply)
	ribdServiceHandler.GlobalPolicyEngineDB.SetTraverseAndReversePolicyFunc(policyEngineTraverseAndReverse)
	ribdServiceHandler.GlobalPolicyEngineDB.SetGetPolicyEntityMapIndexFunc(getPolicyRouteMapIndex)
	return ribdServiceHandler.GlobalPolicyEngineDB
}

func (ribdServiceHandler *RIBDServer) InitializePolicyDB() *policy.PolicyEngineDB {
	ribdServiceHandler.PolicyEngineDB = policy.NewPolicyEngineDB(logger)
	ribdServiceHandler.PolicyEngineDB.SetDefaultImportPolicyActionFunc(defaultImportPolicyEngineActionFunc)
	ribdServiceHandler.PolicyEngineDB.SetDefaultExportPolicyActionFunc(defaultExportPolicyEngineActionFunc)
	ribdServiceHandler.PolicyEngineDB.SetIsEntityPresentFunc(DoesRouteExist)
	ribdServiceHandler.PolicyEngineDB.SetEntityUpdateFunc(UpdateRouteAndPolicyDB)
	ribdServiceHandler.PolicyEngineDB.SetActionFunc(policyCommonDefs.PolicyActionTypeRouteDisposition, policyEngineRouteDispositionAction)
	ribdServiceHandler.PolicyEngineDB.SetActionFunc(policyCommonDefs.PolicyActionTypeRouteRedistribute, policyEngineActionRedistribute)
	ribdServiceHandler.PolicyEngineDB.SetActionFunc(policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise, policyEngineActionNetworkStatementAdvertise)
	ribdServiceHandler.PolicyEngineDB.SetActionFunc(policyCommonDefs.PoilcyActionTypeSetAdminDistance, policyEngineActionSetAdminDistance)
	ribdServiceHandler.PolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PolicyActionTypeRouteDisposition, policyEngineUndoRouteDispositionAction)
	ribdServiceHandler.PolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PolicyActionTypeRouteRedistribute, policyEngineActionUndoRedistribute)
	ribdServiceHandler.PolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PoilcyActionTypeSetAdminDistance, policyEngineActionUndoSetAdminDistance)
	ribdServiceHandler.PolicyEngineDB.SetUndoActionFunc(policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise, policyEngineActionUndoNetworkStatemenAdvertiseAction)
	ribdServiceHandler.PolicyEngineDB.SetTraverseAndApplyPolicyFunc(policyEngineTraverseAndApply)
	ribdServiceHandler.PolicyEngineDB.SetTraverseAndReversePolicyFunc(policyEngineTraverseAndReverse)
	ribdServiceHandler.PolicyEngineDB.SetGetPolicyEntityMapIndexFunc(getPolicyRouteMapIndex)
	return ribdServiceHandler.PolicyEngineDB
}
func NewRIBDServicesHandler(dbHdl *dbutils.DBUtil, loggerC *logging.Writer) *RIBDServer {
	RouteInfoMap = patriciaDB.NewTrie()
	ribdServicesHandler := &RIBDServer{}
	ribdServicesHandler.Logger = loggerC
	logger = loggerC
	localRouteEventsDB = make([]RouteEventInfo, 0)
	RedistributeRouteMap = make(map[string][]RedistributeRouteInfo)
	ribdServicesHandler.Clients = make(map[string]ClientIf)
	TrackReachabilityMap = make(map[string][]string)
	v4routeCreatedTimeMap = make(map[int]string)
	v6routeCreatedTimeMap = make(map[int]string)
	RouteProtocolTypeMapDB = make(map[string]int)
	ReverseRouteProtoTypeMapDB = make(map[int]string)
	ProtocolAdminDistanceMapDB = make(map[string]RouteDistanceConfig)
	PublisherInfoMap = make(map[string]PublisherMapInfo)
	ribdServicesHandler.NextHopInfoMap = make(map[NextHopInfoKey]NextHopInfo)
	ribdServicesHandler.TrackReachabilityCh = make(chan TrackReachabilityInfo, 1000)
	ribdServicesHandler.RouteConfCh = make(chan RIBdServerConfig, 100000)
	ribdServicesHandler.AsicdRouteCh = make(chan RIBdServerConfig, 100000)
	ribdServicesHandler.ArpdRouteCh = make(chan RIBdServerConfig, 5000)
	ribdServicesHandler.NotificationChannel = make(chan NotificationMsg, 5000)
	ribdServicesHandler.PolicyConditionConfCh = make(chan RIBdServerConfig, 5000)
	ribdServicesHandler.PolicyActionConfCh = make(chan RIBdServerConfig, 5000)
	ribdServicesHandler.PolicyStmtConfCh = make(chan RIBdServerConfig, 5000)
	ribdServicesHandler.PolicyDefinitionConfCh = make(chan RIBdServerConfig, 5000)
	ribdServicesHandler.PolicyApplyCh = make(chan ApplyPolicyInfo, 100)
	ribdServicesHandler.PolicyUpdateApplyCh = make(chan ApplyPolicyInfo, 100)
	ribdServicesHandler.DBRouteCh = make(chan RIBdServerConfig, 100000)
	ribdServicesHandler.ServerUpCh = make(chan bool)
	ribdServicesHandler.DBReadDone = make(chan bool)
	ribdServicesHandler.DbHdl = dbHdl
	RouteServiceHandler = ribdServicesHandler
	//ribdServicesHandler.RouteInstallCh = make(chan RouteParams)
	BuildRouteProtocolTypeMapDB()
	BuildProtocolAdminDistanceMapDB()
	BuildPublisherMap()
	PolicyEngineDB = ribdServicesHandler.InitializePolicyDB()
	GlobalPolicyEngineDB = ribdServicesHandler.InitializeGlobalPolicyDB()
	return ribdServicesHandler
}
func (s *RIBDServer) InitServer() {
	sigChan := make(chan os.Signal, 1)
	signalList := []os.Signal{syscall.SIGHUP}
	signal.Notify(sigChan, signalList...)
	go s.ListenToClientStateChanges()
	go s.SigHandler(sigChan)
	go s.StartRouteProcessServer()
	go s.StartDBServer()
	go s.StartPolicyServer()
	go s.NotificationServer()
	go s.StartAsicdServer()
	go s.StartArpdServer()

}
func (ribdServiceHandler *RIBDServer) StartServer(paramsDir string) {
	ribdServiceHandler.InitServer()
	logger.Info("Starting RIB server comment out logger. calls")
	DummyRouteInfoRecord.protocol = PROTOCOL_NONE
	configFile := paramsDir + "/clients.json"
	logger.Info(fmt.Sprintln("configfile = ", configFile))
	PARAMSDIR = paramsDir
	ribdServiceHandler.UpdatePolicyObjectsFromDB() //(paramsDir)
	ribdServiceHandler.ConnectToClients(configFile)
	logger.Info("Starting the server loop")
	count := 0
	for {
		if !RouteServiceHandler.AcceptConfig {
			if count%1000 == 0 {
				logger.Debug("RIBD not ready to accept config")
			}
			count++
			continue
		}
		select {
		case info := <-ribdServiceHandler.PolicyApplyCh:
			//logger.Debug("received message on PolicyApplyCh channel")
			//update the local policyEngineDB
			ribdServiceHandler.UpdateApplyPolicy(info, true, PolicyEngineDB)
			ribdServiceHandler.PolicyUpdateApplyCh <- info
		case info := <-ribdServiceHandler.TrackReachabilityCh:
			//logger.Debug("received message on TrackReachabilityCh channel")
			ribdServiceHandler.TrackReachabilityStatus(info.IpAddr, info.Protocol, info.Op)
		}
	}
}

func (ribdServiceHandler *RIBDServer) SigHandler(sigChan <-chan os.Signal) {
	//logger.Debug("Inside sigHandler....")
	signal := <-sigChan
	switch signal {
	case syscall.SIGHUP:
		//logger.Debug("Received SIGHUP signal")
		//logger.Debug("Closing DB handler")
		if ribdServiceHandler.DbHdl != nil {
			ribdServiceHandler.DbHdl.Disconnect()
		}
		os.Exit(0)
	default:
		//logger.Err(fmt.Sprintln("Unhandled signal : ", signal))
	}
}
