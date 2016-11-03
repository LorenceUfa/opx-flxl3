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
	//"bytes"
	//"encoding/json"
	//"git.apache.org/thrift.git/lib/go/thrift"
	nanomsg "github.com/op/go-nanomsg"
	"l3/rib/ribdCommonDefs"
	"ribd"
	//"ribdInt"
	"strconv"
	"time"
	"utils/ipcutils"
)

type RibdClient struct {
	OspfClientBase
	ClientHdl *ribd.RIBDServicesClient
}

type RibdCommStruct struct {
	ribdSubSocketCh    chan []byte
	ribdClient         RibdClient
	ribdSubSocket      *nanomsg.SubSocket
	ribdSubSocketErrCh chan error
}

func (server *OSPFV2Server) initRibdComm() error {
	server.ribdComm.ribdSubSocketCh = make(chan []byte)
	server.ribdComm.ribdSubSocketErrCh = make(chan error)
	return nil
}

func (server *OSPFV2Server) ConnectToRibdServer(port int) {
	var err error
	server.logger.Info("found ribd at port", port)
	server.ribdComm.ribdClient.Address = "localhost:" + strconv.Itoa(port)
	server.ribdComm.ribdClient.Transport, server.ribdComm.ribdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(server.ribdComm.ribdClient.Address)
	if err != nil {
		server.logger.Info("Failed to connect to ribd, retrying until connection is successful")
		count := 0
		ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
		for _ = range ticker.C {
			server.ribdComm.ribdClient.Transport, server.ribdComm.ribdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(server.ribdComm.ribdClient.Address)
			if err == nil {
				ticker.Stop()
				break
			}
			count++
			if (count % 10) == 0 {
				server.logger.Info("Still can't connect to ribd, retrying..")
			}
		}
	}
	server.logger.Info("Ospfd is connected to ribd")
	server.ribdComm.ribdClient.ClientHdl = ribd.NewRIBDServicesClientFactory(server.ribdComm.ribdClient.Transport, server.ribdComm.ribdClient.PtrProtocolFactory)
	server.ribdComm.ribdClient.IsConnected = true
}

func (server *OSPFV2Server) StartRibdSubscriber() {
	server.logger.Info("Listen for ribd updates")
	server.listenForRibdUpdates(ribdCommonDefs.PUB_SOCKET_ADDR)
	go server.createRibdSubscriber()
}

func (server *OSPFV2Server) listenForRibdUpdates(address string) error {
	var err error
	if server.ribdComm.ribdSubSocket, err = nanomsg.NewSubSocket(); err != nil {
		server.logger.Err("ERR: Failed to create RIB subscribe socket, error:", err)
		return err
	}

	if err = server.ribdComm.ribdSubSocket.Subscribe(""); err != nil {
		server.logger.Err("ERR: Failed to subscribe to \"\" on RIB subscribe socket, error:", err)
		return err
	}

	if _, err = server.ribdComm.ribdSubSocket.Connect(address); err != nil {
		server.logger.Err("ERR: Failed to connect to RIB publisher socket, address:", address, "error:", err)
		return err
	}

	server.logger.Info("Connected to RIB publisher at address:", address)
	if err = server.ribdComm.ribdSubSocket.SetRecvBuffer(1024 * 1024); err != nil {
		server.logger.Err("ERR: Failed to set the buffer size for RIB publisher socket, error:", err)
		return err
	}
	return nil
}

func (server *OSPFV2Server) createRibdSubscriber() {
	for {
		server.logger.Info("Read on Ribd subscriber socket...")
		ribdRxBuf, err := server.ribdComm.ribdSubSocket.Recv(0)
		if err != nil {
			server.logger.Err("ERR: Recv on Ribd subscriber socket failed with error:", err)
			server.ribdComm.ribdSubSocketErrCh <- err
			continue
		}
		server.ribdComm.ribdSubSocketCh <- ribdRxBuf
	}
}

func (server *OSPFV2Server) processRibdNotification(ribdRxBuf []byte) {

}
