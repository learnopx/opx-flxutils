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

// bgpUtils.go
package bgpUtils

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	commondefs "utils/commonDefs"
)

var TypeStrToLowTypeMap = map[string]string{
	"Route-Target": "02",
	"Route-Origin": "03",
}

var ipAddr, asplain, fourbytenn, twobytenn *regexp.Regexp

type ExtCommunity struct {
	Type  string
	Value string
}

func GetHexString(in string, hexlen int) string {
	num, _ := strconv.Atoi(in)
	qualifierStr := "%" + strconv.Itoa(hexlen) + "x"
	val := fmt.Sprintf(qualifierStr, num)
	val = strings.Replace(val, " ", "0", -1)
	return val
}
func GetCommunityValue(inp string) (uint32, error) {
	var val uint32
	info, ok := commondefs.BGPWellKnownCommunitiesMap[inp]
	if ok {
		val = info
	} else if strings.HasPrefix(inp, "0x") {
		info, err := strconv.ParseInt(inp, 0, 64)
		if err != nil {
			return val, err
		} else {
			val = uint32(info)
		}
	} else {
		//split with :
		a := strings.Split(inp, ":")
		if len(a) > 2 {
			fmt.Println("Incorrect format for community ", inp)
			return val, errors.New(fmt.Sprintln("Incorrect format for community ", inp))
		}
		if len(a) == 2 {
			as := GetHexString(a[0], 4)
			num := GetHexString(a[1], 4)
			comm := "0x" + as + num
			valint, err := strconv.ParseInt(comm, 0, 64)
			if err != nil {
				return val, err
			}
			val = uint32(valint)
		} else if len(a) == 1 {
			//just a integer
			info, err := strconv.Atoi(inp)
			//fmt.Println("err:", err, " while caling strconv for ", inp)
			if err == nil {
				val = uint32(info)
			} else {
				fmt.Println("Incorrect community input:", inp)
				return val, err
			}
		} else {
			fmt.Println("Incorrect community input:", inp)
			return val, errors.New(fmt.Sprintln("Incorrect community input:", inp))
		}
	}
	return val, nil
}

func EncodeExtCommunity(inp ExtCommunity) (string, error) {
	ipAddr, _ = regexp.Compile("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$")
	asplain, _ = regexp.Compile("^([0-5]?[0-9]?[0-9]?[0-9]?[0-9]?|6[0-9]?[0-9]?[0-9]?[0-5]?)$")
	twobytenn = asplain
	fourbytenn, _ = regexp.Compile("^([0-3]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?|[4]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-9]?[0-5]?)$")
	//fmt.Println("inp :", inp)
	_, ok := TypeStrToLowTypeMap[inp.Type]
	if !ok {
		fmt.Println("invalid type")
		return "", errors.New("Invalid Type")
	}
	lowByte := ""
	highByte := ""
	value := ""
	lowByte = GetHexString(TypeStrToLowTypeMap[inp.Type], 2)
	comm := "0x"
	a := strings.Split(inp.Value, ":")
	if len(a) != 2 {
		fmt.Println("Invalid ExtCommunity value:", inp.Value)
		return "", errors.New("Invalid Extended community value")
	}

	switch inp.Type {
	case "Route-Target", "Route-Origin":
		if asplain.MatchString(a[0]) && fourbytenn.MatchString(a[1]) { //200:10
			highByte = "00"
			as := GetHexString(a[0], 4)
			fmt.Println("as:", as)
			num := GetHexString(a[1], 8)
			fmt.Println("num:", num)
			value = as + num
		} else if ipAddr.MatchString(a[0]) && twobytenn.MatchString(a[1]) { //255.255.255.255:65535
			highByte = "01"
			num := GetHexString(a[1], 4)
			ip := net.ParseIP(a[0])
			ipBytes := strings.Split(ip.String(), ".") //(net.IPMask(ip)).String()
			if len(ipBytes) == 4 {
				for _, ipByte := range ipBytes {
					ipVal := GetHexString(ipByte, 2)
					value = value + ipVal
				}
				value = value + num
			}
		} else {
			//4byteasn:2byteval 70000:200
			if fourbytenn.MatchString(a[0]) && asplain.MatchString(a[1]) {
				highByte = "02"
				as := GetHexString(a[0], 8)
				num := GetHexString(a[1], 4)
				value = as + num
			} else {
				//asdot format:2 byte asn: 65535.230:310
				a1 := strings.Split(a[0], ".")
				if len(a1) != 2 {
					fmt.Println("Invalid format for extended community")
					return "", errors.New("Invalid Extended community value")
				}
				if twobytenn.MatchString(a1[0]) && twobytenn.MatchString(a1[1]) && twobytenn.MatchString(a[1]) {
					highByte = "02"
					asdot0 := GetHexString(a1[0], 4)
					asdot1 := GetHexString(a1[1], 4)
					num := GetHexString(a[1], 4)
					value = asdot0 + asdot1 + num
				} else {
					fmt.Println("Invalid extended community:", inp.Value)
					return "", errors.New("Invalid Extended community value")
				}
			}
		}
	default:
		fmt.Println("Type:", inp.Type, " not supported")
		return "", errors.New("Invalid Extended community value")
	}
	return comm + highByte + lowByte + value, nil
}
