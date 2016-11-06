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
	"errors"
	"net"
	"strconv"
)

func convertDotNotationToUint32(str string) (uint32, error) {
	var val uint32
	ip := net.ParseIP(str)
	if ip == nil {
		return 0, errors.New("Invalid string format")
	}
	ipBytes := ip.To4()
	val = val + uint32(ipBytes[0])
	val = (val << 8) + uint32(ipBytes[1])
	val = (val << 8) + uint32(ipBytes[2])
	val = (val << 8) + uint32(ipBytes[3])
	return val, nil
}

func convertMaskToUint32(mask net.IPMask) uint32 {
	var val uint32

	val = val + uint32(mask[0])
	val = (val << 8) + uint32(mask[1])
	val = (val << 8) + uint32(mask[2])
	val = (val << 8) + uint32(mask[3])
	return val
}

func ParseCIDRToUint32(IpAddr string) (ip uint32, mask uint32, err error) {
	ipAddr, ipNet, err := net.ParseCIDR(IpAddr)
	if err != nil {
		return 0, 0, errors.New("Invalid IP Address")
	}
	ip, _ = convertDotNotationToUint32(ipAddr.String())
	mask = convertMaskToUint32(ipNet.Mask)
	return ip, mask, nil
}
func convertUint32ToDotNotation(val uint32) string {
	p0 := int(val & 0xFF)
	p1 := int((val >> 8) & 0xFF)
	p2 := int((val >> 16) & 0xFF)
	p3 := int((val >> 24) & 0xFF)
	str := strconv.Itoa(p3) + "." + strconv.Itoa(p2) + "." +
		strconv.Itoa(p1) + "." + strconv.Itoa(p0)

	return str
}
