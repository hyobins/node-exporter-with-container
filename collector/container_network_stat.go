// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nocontainer

package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ContainerNetworkStat struct {
	rxBytes   float64
	rxPackets float64
	txBytes   float64
	txPackets float64
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

//GetNetworkStat returns network stats by container
func getNetworkStat(pid string) ContainerNetworkStat {
	var resultList ContainerNetworkStat

	lines, err := readLines(procFilePath(pid) + "/net/dev")
	if err != nil {
		fmt.Printf("ERROR. Failed to read net/dev. pid[ %s ]", pid)
		panic(err)
	}
	nw := strings.Fields(string(lines[2]))

	rxBytes, _ := strconv.ParseFloat(nw[1], 64)
	rxPackets, _ := strconv.ParseFloat(nw[2], 64)
	txBytes, _ := strconv.ParseFloat(nw[9], 64)
	txPackets, _ := strconv.ParseFloat(nw[10], 64)

	stat := ContainerNetworkStat{
		rxBytes:   rxBytes,
		rxPackets: rxPackets,
		txBytes:   txBytes,
		txPackets: txPackets,
	}

	resultList = stat

	return resultList
}
