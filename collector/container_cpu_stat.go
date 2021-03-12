// Copyright 2019 The Prometheus Authors
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
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type ContainerCPUStat struct {
	pid    string
	utime  float64
	stime  float64
	cutime float64
	cstime float64
}

//GetMemoryStat return cpu usage by container
func getCPUStat(pids []string) map[string]ContainerCPUStat {

	resultList := make(map[string]ContainerCPUStat)

	for i := 0; i < len(pids)-1; i++ {

		// HB
		// cpus, err := (exec.Command("bash", "-c", "cat "+procFilePath(pids[i])+"/stat")).Output()
		// if err != nil {
		// 	panic(err)
		// }
		// Edit JB : 2021.03.11
		cpus, err := ioutil.ReadFile(procFilePath(pids[i] + "/stat"))
		if err != nil {
			fmt.Printf("ERROR. Failed to read file(container[ %s ] cpu stat).[ %s ]",
				pids[i], err.Error())
			panic(err)
		}

		cpu := strings.Split(string(cpus), " ")
		utime, _ := strconv.ParseFloat(cpu[13], 64)
		stime, _ := strconv.ParseFloat(cpu[14], 64)
		cutime, _ := strconv.ParseFloat(cpu[15], 64)
		cstime, _ := strconv.ParseFloat(cpu[16], 64)

		stat := ContainerCPUStat{
			pid:    pids[i],
			utime:  utime,
			stime:  stime,
			cutime: cutime,
			cstime: cstime,
		}

		resultList[pids[i]] = stat

	}

	return resultList
}
