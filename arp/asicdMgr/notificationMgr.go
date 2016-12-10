package asicdMgr

import (
	"l3/arp/server"
	"utils/clntUtils/clntDefs"
	"utils/logging"
)

type NotificationHdl struct {
	Server *server.ARPServer
}

func initAsicdNotification() clntDefs.AsicdNotification {
	nMap := make(clntDefs.AsicdNotification)
	nMap = clntDefs.AsicdNotification{
		clntDefs.NOTIFY_L2INTF_STATE_CHANGE:           true,
		clntDefs.NOTIFY_IPV4_L3INTF_STATE_CHANGE:      true,
		clntDefs.NOTIFY_IPV6_L3INTF_STATE_CHANGE:      false,
		clntDefs.NOTIFY_VLAN_CREATE:                   true,
		clntDefs.NOTIFY_VLAN_DELETE:                   true,
		clntDefs.NOTIFY_VLAN_UPDATE:                   true,
		clntDefs.NOTIFY_LOGICAL_INTF_CREATE:           false,
		clntDefs.NOTIFY_LOGICAL_INTF_DELETE:           false,
		clntDefs.NOTIFY_LOGICAL_INTF_UPDATE:           true,
		clntDefs.NOTIFY_IPV4INTF_CREATE:               true,
		clntDefs.NOTIFY_IPV4INTF_DELETE:               true,
		clntDefs.NOTIFY_IPV6INTF_CREATE:               false,
		clntDefs.NOTIFY_IPV6INTF_DELETE:               false,
		clntDefs.NOTIFY_LAG_CREATE:                    true,
		clntDefs.NOTIFY_LAG_DELETE:                    true,
		clntDefs.NOTIFY_LAG_UPDATE:                    true,
		clntDefs.NOTIFY_IPV4NBR_MAC_MOVE:              true,
		clntDefs.NOTIFY_IPV6NBR_MAC_MOVE:              false,
		clntDefs.NOTIFY_IPV4_ROUTE_CREATE_FAILURE:     false,
		clntDefs.NOTIFY_IPV4_ROUTE_DELETE_FAILURE:     false,
		clntDefs.NOTIFY_IPV6_ROUTE_CREATE_FAILURE:     false,
		clntDefs.NOTIFY_IPV6_ROUTE_DELETE_FAILURE:     false,
		clntDefs.NOTIFY_VTEP_CREATE:                   false,
		clntDefs.NOTIFY_VTEP_DELETE:                   false,
		clntDefs.NOTIFY_MPLSINTF_STATE_CHANGE:         false,
		clntDefs.NOTIFY_MPLSINTF_CREATE:               false,
		clntDefs.NOTIFY_MPLSINTF_DELETE:               false,
		clntDefs.NOTIFY_PORT_CONFIG_MODE_CHANGE:       false,
		clntDefs.NOTIFY_IPV4VIRTUAL_INTF_CREATE:       true,
		clntDefs.NOTIFY_IPV4VIRTUAL_INTF_DELETE:       true,
		clntDefs.NOTIFY_IPV6VIRTUAL_INTF_CREATE:       false,
		clntDefs.NOTIFY_IPV6VIRTUAL_INTF_DELETE:       false,
		clntDefs.NOTIFY_IPV4_VIRTUALINTF_STATE_CHANGE: true,
		clntDefs.NOTIFY_IPV6_VIRTUALINTF_STATE_CHANGE: false,
	}
	return nMap
}

func NewNotificationHdl(server *server.ARPServer, logger *logging.Writer) (clntDefs.AsicdNotificationHdl, clntDefs.AsicdNotification) {
	nMap := initAsicdNotification()
	return &NotificationHdl{server}, nMap
}

func (nHdl *NotificationHdl) ProcessNotification(msg clntDefs.AsicdNotifyMsg) {
	nHdl.Server.AsicdSubSocketCh <- msg
}
