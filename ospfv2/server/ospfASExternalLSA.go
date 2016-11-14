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
//"errors"
//"fmt"
//"l3/ospfv2/objects"
)

func (server *OSPFV2Server) processRecvdSelfASExternalLSA(msg RecvdSelfLsaMsg) error {
	/*
	   lsa, ok := msg.LsaData.(RouterLsa)
	   if !ok {
	           server.logger.Err("Unable to assert given router lsa")
	           return nil
	   }
	   lsdbEnt, exist := server.LsdbData.AreaLsdb[msg.LsdbKey]
	   if !exist {
	           server.logger.Err("No such Area exist", msg.LsdbKey)
	           return nil
	   }
	   lsaEnt, exist := lsdbEnt.RouterLsaMap[msg.LsaKey]
	   if !exist {
	           server.logger.Err("No such router LSA exist", msg.LsaKey)
	           return nil
	   }
	   selfOrigLsaEnt, exist := server.LsdbData.AreaSelfOrigLsa[msg.LsdbKey]
	   if !exist {
	           server.logger.Err("No such self originated router LSA exist")
	           return nil
	   }
	   if lsaEnt.LsaMd.LSSequenceNum > lsa.LsaMd.LSSequenceNum {
	           lsaEnt.LsaMd.LSSequenceNum = lsa.LsaMd.LSSequenceNum + 1
	           lsaEnt.LsaMd.LSAge = 0
	           lsaEnc := encodeRouterLsa(lsaEnt, msg.LsaKey)
	           lsaEnt.LsaMd.LSChecksum = computeFletcherChecksum(lsaEnc[2:], checksumOffset)
	           lsdbEnt.RouterLsaMap[msg.LsaKey] = lsaEnt
	           server.LsdbData.AreaLsdb[msg.LsdbKey] = lsdbEnt
	           //TODO:
	           //Flood new Self Router LSA
	           return nil
	   } else {
	           //TODO:
	           //Flood existing Self Router LSA
	   }
	*/

	return nil
}

func (server *OSPFV2Server) processRecvdASExternalLSA(msg RecvdLsaMsg) error {
	/*
	   lsdbEnt, exist := server.LsdbData.AreaLsdb[msg.LsdbKey]
	   if !exist {
	           server.logger.Err("No such Area exist", msg.LsdbKey)
	           return nil
	   }
	   if msg.MsgType == LSA_ADD {
	           lsa, ok := msg.LsaData.(RouterLsa)
	           if !ok {
	                   server.logger.Err("Unable to assert given router lsa")
	                   return nil
	           }
	           lsdbEnt.RouterLsaMap[msg.LsaKey] = lsa
	   } else if msg.MsgType == LSA_DEL {
	           delete(lsdbEnt.RouterLsaMap, msg.LsaKey)
	   }
	   server.LsdbData.AreaLsdb[msg.LsdbKey] = lsdbEnt
	*/
	return nil
}
