package asicdMgr

import (
	"l3/arp/server"
	"utils/clntUtils/clntDefs/asicdClntDefs"
	"utils/logging"
)

type NotificationHdl struct {
	Server *server.ARPServer
}

func initAsicdNotification() asicdClntDefs.AsicdNotification {
	nMap := make(asicdClntDefs.AsicdNotification)
	nMap = asicdClntDefs.AsicdNotification{
		asicdClntDefs.NOTIFY_L2INTF_STATE_CHANGE:           true,
		asicdClntDefs.NOTIFY_IPV4_L3INTF_STATE_CHANGE:      true,
		asicdClntDefs.NOTIFY_IPV6_L3INTF_STATE_CHANGE:      false,
		asicdClntDefs.NOTIFY_VLAN_CREATE:                   true,
		asicdClntDefs.NOTIFY_VLAN_DELETE:                   true,
		asicdClntDefs.NOTIFY_VLAN_UPDATE:                   true,
		asicdClntDefs.NOTIFY_LOGICAL_INTF_CREATE:           false,
		asicdClntDefs.NOTIFY_LOGICAL_INTF_DELETE:           false,
		asicdClntDefs.NOTIFY_LOGICAL_INTF_UPDATE:           true,
		asicdClntDefs.NOTIFY_IPV4INTF_CREATE:               true,
		asicdClntDefs.NOTIFY_IPV4INTF_DELETE:               true,
		asicdClntDefs.NOTIFY_IPV6INTF_CREATE:               false,
		asicdClntDefs.NOTIFY_IPV6INTF_DELETE:               false,
		asicdClntDefs.NOTIFY_LAG_CREATE:                    true,
		asicdClntDefs.NOTIFY_LAG_DELETE:                    true,
		asicdClntDefs.NOTIFY_LAG_UPDATE:                    true,
		asicdClntDefs.NOTIFY_IPV4NBR_MAC_MOVE:              true,
		asicdClntDefs.NOTIFY_IPV6NBR_MAC_MOVE:              false,
		asicdClntDefs.NOTIFY_IPV4_ROUTE_CREATE_FAILURE:     false,
		asicdClntDefs.NOTIFY_IPV4_ROUTE_DELETE_FAILURE:     false,
		asicdClntDefs.NOTIFY_IPV6_ROUTE_CREATE_FAILURE:     false,
		asicdClntDefs.NOTIFY_IPV6_ROUTE_DELETE_FAILURE:     false,
		asicdClntDefs.NOTIFY_VTEP_CREATE:                   false,
		asicdClntDefs.NOTIFY_VTEP_DELETE:                   false,
		asicdClntDefs.NOTIFY_MPLSINTF_STATE_CHANGE:         false,
		asicdClntDefs.NOTIFY_MPLSINTF_CREATE:               false,
		asicdClntDefs.NOTIFY_MPLSINTF_DELETE:               false,
		asicdClntDefs.NOTIFY_PORT_CONFIG_MODE_CHANGE:       false,
		asicdClntDefs.NOTIFY_IPV4VIRTUAL_INTF_CREATE:       true,
		asicdClntDefs.NOTIFY_IPV4VIRTUAL_INTF_DELETE:       true,
		asicdClntDefs.NOTIFY_IPV6VIRTUAL_INTF_CREATE:       false,
		asicdClntDefs.NOTIFY_IPV6VIRTUAL_INTF_DELETE:       false,
		asicdClntDefs.NOTIFY_IPV4_VIRTUALINTF_STATE_CHANGE: true,
		asicdClntDefs.NOTIFY_IPV6_VIRTUALINTF_STATE_CHANGE: false,
	}
	return nMap
}

func NewNotificationHdl(server *server.ARPServer, logger *logging.Writer) (asicdClntDefs.AsicdNotificationHdl, asicdClntDefs.AsicdNotification) {
	nMap := initAsicdNotification()
	return &NotificationHdl{server}, nMap
}

func (nHdl *NotificationHdl) ProcessNotification(msg asicdClntDefs.AsicdNotifyMsg) {
	nHdl.Server.AsicdSubSocketCh <- msg
}
