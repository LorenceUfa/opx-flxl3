// ribdAsicdServer.go
package main

import (
)
func addAsicdRoute(routeInfoRecord RouteInfoRecord) {
	logger.Info("addAsicdRoute")
	asicdclnt.ClientHdl.OnewayCreateIPv4Route(routeInfoRecord.destNetIp.String(), routeInfoRecord.networkMask.String(), routeInfoRecord.resolvedNextHopIpIntf.NextHopIp, int32(routeInfoRecord.resolvedNextHopIpIntf.NextHopIfType))
}
func delAsicdRoute(routeInfoRecord RouteInfoRecord) {
	logger.Info("delAsicdRoute")
	asicdclnt.ClientHdl.OnewayDeleteIPv4Route(routeInfoRecord.destNetIp.String(), routeInfoRecord.networkMask.String(), routeInfoRecord.resolvedNextHopIpIntf.NextHopIp, int32(routeInfoRecord.resolvedNextHopIpIntf.NextHopIfType))
}
func (ribdServiceHandler *RIBDServicesHandler) StartAsicdServer() {
	logger.Info("Starting the asicdserver loop")
	for {
		select {
		case route := <-ribdServiceHandler.AsicdAddRouteCh:
		     logger.Info(" received message on AsicdAddRouteCh")
		     addAsicdRoute(route)
		case route := <-ribdServiceHandler.AsicdDelRouteCh:
		     logger.Info(" received message on AsicdDelRouteCh")
		     delAsicdRoute(route)
		}
	}
}