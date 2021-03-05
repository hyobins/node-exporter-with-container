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
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
)

type ContainerCpuStat struct {
	pid           string
	containerID   string
	containerName string
	utime         float64
	stime         float64
	cutime        float64
	cstime        float64
	//	value 		  ValuebyPid
}

type ValuebyPid struct {
	pid    int64
	utime  float64
	stime  float64
	cutie  float64
	cstime float64
}

//GetMemoryStat return container's cpu usage
func getCpuStat(containers []types.Container) map[string]ContainerCpuStat {

	resultList := make(map[string]ContainerCpuStat)
	stat := ContainerCpuStat{} //initialize

	for _, container := range containers {
		m := make(map[string]interface{})
		//v := ValuebyPid{} //initialize e

		pidpath := exec.Command("bash", "-c", "cd /run/docker/runtime-runc/moby/"+container.ID+" && cat state.json")
		outputPath, _ := pidpath.Output()

		//parse state.json of each container
		err := json.Unmarshal(outputPath, &m)
		if err != nil {
			panic(err)
		}

		jsondata, _ := json.Marshal(m["cgroup_paths"].(map[string]interface{})["pids"])

		//Get PIDs of each container
		pid, _ := (exec.Command("bash", "-c", "cd "+string(jsondata)+" && cat tasks")).Output()

		slice := strings.Split(string(pid), "\n") //컨테이너별 pid 목록을 slice에 하나씩 저장

		for i := 0; i < len(slice); i++ {
			cpus, err := (exec.Command("bash", "-c", "cat /proc/"+slice[i]+"/stat")).Output()
			if err != nil {
				panic(err)
			}

			cpu := strings.Split(string(cpus), " ")
			//pid, _ := strconv.ParseInt(slice[i], 0, 64)
			utime, _ := strconv.ParseFloat(cpu[13], 64)
			stime, _ := strconv.ParseFloat(cpu[14], 64)
			cutime, _ := strconv.ParseFloat(cpu[15], 64)
			cstime, _ := strconv.ParseFloat(cpu[16], 64)

			//v = ValuebyPid{
			//	pid: pid,
			//	utime: utime,
			//	stime: stime,
			//	cutime: cutime,
			//	cstime: cstime,
			//}

			containerName := strings.Join(container.Names, "")

			stat = ContainerCpuStat{
				pid:           slice[i],
				containerID:   container.ID,
				containerName: containerName[1:],
				utime:         utime,
				stime:         stime,
				cutime:        cutime,
				cstime:        cstime,
			}

			resultList[slice[i]] = stat

		}
	}

	fmt.Println(resultList)

	return resultList
}
