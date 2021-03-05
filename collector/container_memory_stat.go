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
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
)

type ContainerMemoryStat struct {
	pid           string
	containerID   string
	containerName string
	VmSize        float64
	VmRss         float64
	RssFile       float64
}

//GetMemoryStat return container's cpu usage
func getMemoryStat(containers []types.Container) map[string]ContainerMemoryStat {

	resultList := make(map[string]ContainerMemoryStat)
	stat := ContainerMemoryStat{} //initialize

	for _, container := range containers {
		m := make(map[string]interface{})

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
		slice := strings.Split(string(pid), "\n")
		fmt.Printf("slice: %s, len: %d", slice, len(slice))

		for i := 0; i < len(slice)-1; i++ {
			fmt.Println("\nslice[i]:", slice[i])
			lines, err := readLines("/proc/" + slice[i] + "/status")

			if err != nil {
				panic(err)
			}
			re := regexp.MustCompile("[0-9]+")
			m1 := re.FindAllString(string(lines[13]), -1)
			m2 := re.FindAllString(string(lines[17]), -1)
			m3 := re.FindAllString(string(lines[19]), -1)

			vmsize, _ := strconv.ParseFloat(strings.Join(m1, ""), 64)
			vmrss, _ := strconv.ParseFloat(strings.Join(m2, ""), 64)
			rssfile, _ := strconv.ParseFloat(strings.Join(m3, ""), 64)

			containerName := strings.Join(container.Names, "")

			stat = ContainerMemoryStat{
				pid:           slice[i],
				containerID:   container.ID,
				containerName: containerName[1:],
				VmSize:        vmsize,
				VmRss:         vmrss,
				RssFile:       rssfile,
			}

			resultList[slice[i]] = stat

		}

	}
	fmt.Println("메모리 리스트", resultList)

	return resultList
}
