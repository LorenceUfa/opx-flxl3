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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"ribd"
	"ribdInt"
	"strings"
	"utils/patriciaDB"
	"utils/policy"
	"utils/policy/policyCommonDefs"
)

/*
   Function to create policy prefix set in the policyEngineDB
*/
func (m RIBDServer) ProcessPolicyASPathSetConfigCreate(cfg *ribd.PolicyASPathSet, db *policy.PolicyEngineDB) (val bool, err error) {
	logger.Debug("ProcessPolicyASPathSetConfigCreate:", cfg.Name)
	list := make([]string, 0)
	for _, v := range cfg.ASPathList {
		list = append(list, v.ASPath)
	}
	newCfg := policy.PolicyASPathSetConfig{Name: cfg.Name, ASPathList: list}
	val, err = db.CreatePolicyASPathSet(newCfg)
	return val, err
}

/*
   Function to delete policy prefix set in the policyEngineDB
*/
func (m RIBDServer) ProcessPolicyASPathSetConfigDelete(cfg *ribd.PolicyASPathSet, db *policy.PolicyEngineDB) (val bool, err error) {
	logger.Debug("ProcessPolicyASPathSetConfigDelete: ", cfg.Name)
	newCfg := policy.PolicyASPathSetConfig{Name: cfg.Name}
	val, err = db.DeletePolicyASPathSet(newCfg)
	return val, err
}

/*
   Function to patch update policy prefix set in the policyEngineDB
*/
func (m RIBDServer) ProcessPolicyASPathSetConfigPatchUpdate(origCfg *ribd.PolicyASPathSet, newCfg *ribd.PolicyASPathSet, op []*ribd.PatchOpInfo, db *policy.PolicyEngineDB) (err error) {
	logger.Debug("ProcessPolicyASPathSetConfigUpdate:", origCfg.Name)
	if origCfg.Name != newCfg.Name {
		logger.Err("Update for a different policy aspath set")
		return errors.New("Policy aspath set to be updated is different than the original one")
	}
	for idx := 0; idx < len(op); idx++ {
		switch op[idx].Path {
		case "ASPathList":
			logger.Debug("Patch update for ASPathList")
			newPolicyObj := policy.PolicyASPathSetConfig{
				Name: origCfg.Name,
			}
			newPolicyObj.ASPathList = make([]string, 0)
			valueObjArr := []string
			err = json.Unmarshal([]byte(op[idx].Value), &valueObjArr)
			if err != nil {
				//logger.Debug("error unmarshaling value:", err))
				return errors.New(fmt.Sprintln("error unmarshaling value:", err))
			}
			logger.Debug("Number of aspaths:", len(valueObjArr))
			for _, val := range valueObjArr {
				newPolicyObj.ASPathList = append(newPolicyObj.ASPathList, val)
			}
			switch op[idx].Op {
			case "add":
				//db.UpdateAddPolicyDefinition(newPolicy)
			case "remove":
				//db.UpdateRemovePolicyDefinition(newconfig)
			default:
				logger.Err("Operation ", op[idx].Op, " not supported")
			}
		default:
			logger.Err("Patch update for attribute:", op[idx].Path, " not supported")
			err = errors.New(fmt.Sprintln("Operation ", op[idx].Op, " not supported"))
		}
	}
	return err
}

/*
   Function to update policy prefix set in the policyEngineDB
*/
func (m RIBDServer) ProcessPolicyASPathSetConfigUpdate(origCfg *ribd.PolicyASPathSet, newCfg *ribd.PolicyASPathSet, attrset []bool, db *policy.PolicyEngineDB) (err error) {
	logger.Debug("ProcessPolicyASPathSetConfigUpdate:", origCfg.Name)
	if origCfg.Name != newCfg.Name {
		logger.Err("Update for a different policy aspath set statement")
		return errors.New("Policy aspath set statement to be updated is different than the original one")
	}
	return err
}
