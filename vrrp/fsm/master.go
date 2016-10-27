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

package fsm

import (
	"bytes"
	"l3/vrrp/debug"
	"net"
	"time"
)

func (f *FSM) TransitionToMaster() {
	pktInfo := f.getPacketInfo()
	// (110) + Send an ADVERTISEMENT
	f.SendPkt(pktInfo)
	// (145) + Transition to the {Master} state
	f.State = VRRP_MASTER_STATE
	// @TODO : Set Sub-intf state up and send out garp via linux stack
	// svr.VrrpUpdateSubIntf(gblInfo, true /*configure or set*/ //)

	// (140) + Set the Adver_Timer to Advertisement_Interval
	// Start Advertisement Timer
	f.StartMasterAdverTimer()
}

func (f *FSM) StartMasterAdverTimer() {
	if f.AdverTimer != nil {
		f.AdverTimer.Reset(time.Duration(f.Config.AdvertisementInterval) * time.Second)
	} else {
		var SendMasterAdveristement_func func()
		SendMasterAdveristement_func = func() {
			// Send advertisment every time interval expiration
			f.SendPkt(f.getPacketInfo())
			f.AdverTimer.Reset(time.Duration(f.Config.AdvertisementInterval) * time.Second)
		}
		debug.Logger.Debug("Setting Master Advertisement Timer to:", f.Config.AdvertisementInterval)
		f.AdverTimer = time.AfterFunc(time.Duration(f.Config.AdvertisementInterval), SendMasterAdveristement_func)
	}
}

func (f *FSM) StopMasterAdverTimer() {
	if f.AdverTimer != nil {
		f.AdverTimer.Stop()
		f.AdverTimer = nil
	}
}

func (f *FSM) StopMasterDownTimer() {
	if f.MasterDownTimer != nil {
		f.MasterDownTimer.Stop()
		f.MasterDownTimer = nil
	}
}

func (f *FSM) HandleMasterDownTimer() {
	if f.MasterDownTimer != nil {
		f.MasterDownTimer.Reset(time.Duration(f.MasterDownValue) * time.Second)
	} else {
		var MasterDownTimer_func func()
		// On Timer expiration we will transition to master
		MasterDownTimer_func = func() {
			debug.Logger.Info(FSM_PREFIX, "master down timer expired..transition to Master")
			f.TransitionToMaster()
		}
		debug.Logger.Info("setting down timer to", f.MasterDownValue)
		// Set Timer expire func...
		f.MasterDownTimer = time.AfterFunc(time.Duration(f.MasterDownValue)*time.Second, MasterDownTimer_func)
	}
}

func (f *FSM) MasterState(stInfo *FsmStateInfo) {
	debug.Logger.Debug(FSM_PREFIX, "In Master State Handling Fsm Info:", *stInfo)
	pktInfo := stInfo.PktInfo
	hdr := pktInfo.Hdr
	/* // @TODO:
	   (645) - MUST forward packets with a destination link-layer MAC
	   address equal to the virtual router MAC address.

	   (650) - MUST accept packets addressed to the IPvX address(es)
	   associated with the virtual router if it is the IPvX address owner
	   or if Accept_Mode is True.  Otherwise, MUST NOT accept these
	   packets.
	*/
	//  (700) - If an ADVERTISEMENT is received, then:
	//	 (705) -+ If the Priority in the ADVERTISEMENT is zero, then:
	if hdr.Priority == VRRP_MASTER_DOWN_PRIORITY {
		// (710) -* Send an ADVERTISEMENT
		debug.Logger.Debug(FSM_PREFIX, "Priority in the ADVERTISEMENT is zero, then: Send an ADVERTISEMENT")
		f.SendPkt(f.getPacketInfo())
		// (715) -* Reset the Adver_Timer to Advertisement_Interval
		f.StartMasterAdverTimer()
	} else { // (720) -+ else // priority was non-zero
		/*     (725) -* If the Priority in the ADVERTISEMENT is greater than the local Priority,
		*      (730) -* or
		*      (735) -* If the Priority in the ADVERTISEMENT is equal to
		*               the local Priority and the primary IPvX Address of the
		*	        sender is greater than the local primary IPvX Address, then:
		 */
		if int32(hdr.Priority) > f.Config.Priority ||
			(int32(hdr.Priority) == f.Config.Priority &&
				bytes.Compare(net.ParseIP(pktInfo.IpAddr), net.ParseIP(f.IpAddr)) > 0) {
			// (740) -@ Cancel Adver_Timer
			f.StopMasterAdverTimer()
			/*
				(745) -@ Set Master_Adver_Interval to Adver Interval contained in the ADVERTISEMENT
				(750) -@ Recompute the Skew_Time
				(755) @ Recompute the Master_Down_Interval
				(760) @ Set Master_Down_Timer to Master_Down_Interval
				(765) @ Transition to the {Backup} state
			*/
			f.TransitionToBackup(int32(hdr.MaxAdverInt))
		} else { // new Master logic
			// Discard Advertisement
			return
		} // endif new Master Detected
	} // end if was priority zero
	// end for Advertisemtn received over the channel
	// end MASTER STATE
}
