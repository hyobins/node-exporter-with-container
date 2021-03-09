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
	"regexp"
	"strconv"
	"strings"
)

type ContainerMemoryStat struct {
	pid     string
	VmSize  float64
	VmRss   float64
	RssFile float64
}

//GetMemoryStat return memory usage by container
func getMemoryStat(pids []string) map[string]ContainerMemoryStat {

	resultList := make(map[string]ContainerMemoryStat)
	stat := ContainerMemoryStat{} //initialize

	for i := 0; i < len(pids)-1; i++ {
		lines, err := readLines("/proc/" + pids[i] + "/status")

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

		stat = ContainerMemoryStat{
			pid:     pids[i],
			VmSize:  vmsize,
			VmRss:   vmrss,
			RssFile: rssfile,
		}

		resultList[pids[i]] = stat

	}

	return resultList
}
