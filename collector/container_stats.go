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

	rx_bytes   *prometheus.Desc
	rx_packets *prometheus.Desc
	tx_bytes   *prometheus.Desc
	tx_packets *prometheus.Desc
}

func init() {
	registerCollector("container", defaultEnabled, NewContainerCollector)
}

//NewContainerCollector returns container stats
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
		rx_bytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_rxbytes"),
			"Current Network Receiver bytes by container.",
			[]string{"id", "name", "type"}, nil,
		),
		rx_packets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_rxpackets"),
			"Current Network Receiver packets by container.",
			[]string{"id", "name", "type"}, nil,
		),
		tx_bytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_net_txbytes"),
			"Current Network Transmitter bytes by container.",
			[]string{"id", "name", "type"}, nil,
		),
		tx_packets: prometheus.NewDesc(
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
		panic(err)
	}

	cpuStatList := getCpuStat(containers)
	for id, list := range cpuStatList {
		ch <- prometheus.MustNewConstMetric(
			c.utime,
			prometheus.GaugeValue, list.utime, list.containerID, list.containerName, "utime", id,
		)
		ch <- prometheus.MustNewConstMetric(
			c.stime,
			prometheus.GaugeValue, list.stime, list.containerID, list.containerName, "stime", id,
		)
		ch <- prometheus.MustNewConstMetric(
			c.cutime,
			prometheus.GaugeValue, list.cutime, list.containerID, list.containerName, "cutime", id,
		)
		ch <- prometheus.MustNewConstMetric(
			c.cstime,
			prometheus.GaugeValue, list.cstime, list.containerID, list.containerName, "cstime", id,
		)
	}

	memoryStatList := getMemoryStat(containers)
	for id, list := range memoryStatList {
		ch <- prometheus.MustNewConstMetric(
			c.vmsize,
			prometheus.GaugeValue, list.VmSize, list.containerID, list.containerName, "vmsize", id,
		)
		ch <- prometheus.MustNewConstMetric(
			c.vmrss,
			prometheus.GaugeValue, list.VmRss, list.containerID, list.containerName, "vmrss", id,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rssfile,
			prometheus.GaugeValue, list.RssFile, list.containerID, list.containerName, "rssfile", id,
		)
	}

	networkStatList := getNetworkStat(containers)
	for id, list := range networkStatList {
		ch <- prometheus.MustNewConstMetric(
			c.rx_bytes,
			prometheus.GaugeValue, list.rx_bytes, id, list.containerName, "rx_bytes",
		)
		ch <- prometheus.MustNewConstMetric(
			c.rx_packets,
			prometheus.GaugeValue, list.rx_packets, id, list.containerName, "rx_packets",
		)
		ch <- prometheus.MustNewConstMetric(
			c.tx_bytes,
			prometheus.GaugeValue, list.tx_bytes, id, list.containerName, "tx_bytes",
		)
		ch <- prometheus.MustNewConstMetric(
			c.tx_packets,
			prometheus.GaugeValue, list.tx_packets, id, list.containerName, "tx_packets",
		)

	}
	return nil

}
