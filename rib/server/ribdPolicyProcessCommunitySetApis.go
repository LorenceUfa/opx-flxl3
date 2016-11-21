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
func (m RIBDServer) ProcessPolicyCommunitySetConfigCreate(cfg *ribd.PolicyCommunitySet, db *policy.PolicyEngineDB) (val bool, err error) {
	logger.Debug("ProcessPolicyCommunitySetConfigCreate:", cfg.Name)
	list := make([]string, 0)
	for _, v := range cfg.CommunityList {
		list = append(list, v.Community)
	}
	newCfg := policy.PolicyCommunitySetConfig{Name: cfg.Name, CommunityList: list}
	val, err = db.CreatePolicyCommunitySet(newCfg)
	return val, err
}

/*
   Function to delete policy prefix set in the policyEngineDB
*/
func (m RIBDServer) ProcessPolicyCommunitySetConfigDelete(cfg *ribd.PolicyCommunitySet, db *policy.PolicyEngineDB) (val bool, err error) {
	logger.Debug("ProcessPolicyCommunitySetConfigDelete: ", cfg.Name)
	newCfg := policy.PolicyCommunitySetConfig{Name: cfg.Name}
	val, err = db.DeletePolicyCommunitySet(newCfg)
	return val, err
}

/*
   Function to patch update policy prefix set in the policyEngineDB
*/
func (m RIBDServer) ProcessPolicyCommunitySetConfigPatchUpdate(origCfg *ribd.PolicyCommunitySet, newCfg *ribd.PolicyCommunitySet, op []*ribd.PatchOpInfo, db *policy.PolicyEngineDB) (err error) {
	logger.Debug("ProcessPolicyCommunitySetConfigUpdate:", origCfg.Name)
	if origCfg.Name != newCfg.Name {
		logger.Err("Update for a different policy community set")
		return errors.New("Policy community set to be updated is different than the original one")
	}
	for idx := 0; idx < len(op); idx++ {
		switch op[idx].Path {
		case "CommunityList":
			logger.Debug("Patch update for CommunityList")
			newPolicyObj := policy.PolicyCommunitySetConfig{
				Name: origCfg.Name,
			}
			newPolicyObj.CommunityList = make([]string, 0)
			valueObjArr := []string
			err = json.Unmarshal([]byte(op[idx].Value), &valueObjArr)
			if err != nil {
				//logger.Debug("error unmarshaling value:", err))
				return errors.New(fmt.Sprintln("error unmarshaling value:", err))
			}
			logger.Debug("Number of communities:", len(valueObjArr))
			for _, val := range valueObjArr {
				newPolicyObj.CommunityList = append(newPolicyObj.CommunityList, val)
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
func (m RIBDServer) ProcessPolicyCommunitySetConfigUpdate(origCfg *ribd.PolicyCommunitySet, newCfg *ribd.PolicyCommunitySet, attrset []bool, db *policy.PolicyEngineDB) (err error) {
	logger.Debug("ProcessPolicyCommunitySetConfigUpdate:", origCfg.Name)
	if origCfg.Name != newCfg.Name {
		logger.Err("Update for a different policy community set statement")
		return errors.New("Policy community set statement to be updated is different than the original one")
	}
	return err
}
