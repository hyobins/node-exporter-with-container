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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-kit/kit/log"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type containerCollector struct {
	logger log.Logger
	utime  *prometheus.Desc
	stime  *prometheus.Desc
	cutime *prometheus.Desc
	cstime *prometheus.Desc

	vmsize  *prometheus.Desc
	vmrss   *prometheus.Desc
	rssfile *prometheus.Desc

	rxBytes   *prometheus.Desc
	rxPackets *prometheus.Desc
	txBytes   *prometheus.Desc
	txPackets *prometheus.Desc
}

func init() {
	registerCollector("container", defaultEnabled, NewContainerCollector)
}

//NewContainerCollector returns a collector exposing hardware resource stats by container
func NewContainerCollector(logger log.Logger) (Collector, error) {
	return &containerCollector{
		logger: logger,
		utime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_cpu_utime"),
			"Current CPU utime by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		stime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_cpu_stime"),
			"Current CPU stime by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		cutime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_cpu_cutime"),
			"Current CPU cutime by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		cstime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_cpu_cstime"),
			"Current CPU cstime by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		vmsize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_memory_vmsize"),
			"Current Memory VmSize by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		vmrss: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_memory_vmrss"),
			"Current Memory VmRss by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		rssfile: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_memory_rssfile"),
			"Current Memory RssFile by container.",
			[]string{"id", "name", "type", "pid"}, nil,
		),
		rxBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_rxbytes"),
			"Current Network Receiver bytes by container.",
			[]string{"id", "name", "type"}, nil,
		),
		rxPackets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_rxpackets"),
			"Current Network Receiver packets by container.",
			[]string{"id", "name", "type"}, nil,
		),
		txBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_txbytes"),
			"Current Network Transmitter bytes by container.",
			[]string{"id", "name", "type"}, nil,
		),
		txPackets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_txpackets"),
			"Current Network Transmitter packets by container.",
			[]string{"id", "name", "type"}, nil,
		),
	}, nil
}

func (c *containerCollector) Update(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		// Edit : JB. 2021.03.11
		// panic(err)
		return fmt.Errorf("ERROR. Failed To Read Container List. [ %s ]", err.Error())
	}

	for _, container := range containers {

		// Edit : JB. 2021.03.11
		// Read state.json
		statePath := fmt.Sprintf("%s/%s/state.json", runFilePath("docker/runtime-runc/moby"), container.ID)
		stateFile, err := ioutil.ReadFile(statePath)
		if err != nil {
			return fmt.Errorf("ERROR. Failed to read file(container[ %s / %+v] state.json)[ %s ]",
				container.ID, container.Names, err.Error())
		}
		// Decode state.json
		stateJsonData := make(map[string]interface{})
		err = json.Unmarshal(stateFile, &stateJsonData)
		if err != nil {
			return fmt.Errorf("ERROR. Failed to decode state.json[ %s ]", err.Error())
		}
		// Get pidfile path from state.json
		cgroupPaths, isExist := stateJsonData["cgroup_paths"].(map[string]interface{})
		if !isExist {
			return fmt.Errorf("ERROR. Not found cgroup_paths from state.json")
		}
		pidPath, isExist := cgroupPaths["pids"].(string)
		if !isExist {
			return fmt.Errorf("ERROR. Not found cgroup_paths - pids from state.json")
		}
		// delete "/sys" from pidPath
		pidPath = strings.TrimLeft(pidPath, "/sys")
		// join runfs + pidpath
		pidFilePath := fmt.Sprintf("%s/tasks", sysFilePath(pidPath))
		// read container pids(tasks)
		pid, err := ioutil.ReadFile(pidFilePath)
		if err != nil {
			return fmt.Errorf("ERROR. Failed to read file(container[ %s / %+v] pid).[ %s ]",
				container.ID, container.Names, err.Error())
		}

		// HB
		// m := make(map[string]interface{})
		// pidpath := exec.Command("bash", "-c", "cd "+runFilePath("docker/runtime-runc/moby")+"/"+container.ID+" && cat state.json")
		// outputPath, _ := pidpath.Output()

		// err := json.Unmarshal(outputPath, &m)
		// if err != nil {
		// 	panic(err)
		// }
		// jsondata, _ := json.Marshal(m["cgroup_paths"].(map[string]interface{})["pids"])
		// pid, _ := (exec.Command("bash", "-c", "cd "+string(jsondata)+" && cat tasks")).Output()

		slice := strings.Split(string(pid), "\n")

		containerName := strings.Join(container.Names, "")

		cpuStatList := getCPUStat(slice)
		for id, list := range cpuStatList {
			ch <- prometheus.MustNewConstMetric(
				c.utime,
				prometheus.GaugeValue, list.utime, container.ID, containerName, "utime", id,
			)
			ch <- prometheus.MustNewConstMetric(
				c.stime,
				prometheus.GaugeValue, list.stime, container.ID, containerName, "stime", id,
			)
			ch <- prometheus.MustNewConstMetric(
				c.cutime,
				prometheus.GaugeValue, list.cutime, container.ID, containerName, "cutime", id,
			)
			ch <- prometheus.MustNewConstMetric(
				c.cstime,
				prometheus.GaugeValue, list.cstime, container.ID, containerName, "cstime", id,
			)
		}

		memoryStatList := getMemoryStat(slice)
		for id, list := range memoryStatList {
			ch <- prometheus.MustNewConstMetric(
				c.vmsize,
				prometheus.GaugeValue, list.VMSize, container.ID, containerName, "vmsize", id,
			)
			ch <- prometheus.MustNewConstMetric(
				c.vmrss,
				prometheus.GaugeValue, list.VMRss, container.ID, containerName, "vmrss", id,
			)
			ch <- prometheus.MustNewConstMetric(
				c.rssfile,
				prometheus.GaugeValue, list.RssFile, container.ID, containerName, "rssfile", id,
			)
		}

		networkStatList := getNetworkStat(slice[0])
		ch <- prometheus.MustNewConstMetric(
			c.rxBytes,
			prometheus.GaugeValue, networkStatList.rxBytes, container.ID, containerName, "rx_bytes",
		)
		ch <- prometheus.MustNewConstMetric(
			c.rxPackets,
			prometheus.GaugeValue, networkStatList.rxPackets, container.ID, containerName, "rx_packets",
		)
		ch <- prometheus.MustNewConstMetric(
			c.txBytes,
			prometheus.GaugeValue, networkStatList.txBytes, container.ID, containerName, "tx_bytes",
		)
		ch <- prometheus.MustNewConstMetric(
			c.txPackets,
			prometheus.GaugeValue, networkStatList.txPackets, container.ID, containerName, "tx_packets",
		)

	}
	return nil

}
