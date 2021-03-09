package collector

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type ContainerNetworkStat struct {
	containerID   string
	containerName string
	rx_bytes      float64
	rx_packets    float64
	tx_bytes      float64
	tx_packets    float64
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
	//resultList := make(map[string]ContainerNetworkStat)
	stat := ContainerNetworkStat{} //initialize

	lines, err := readLines("/proc/" + pid + "/net/dev")
	if err != nil {
		panic(err)
	}
	nw := strings.Fields(string(lines[2]))

	rx_bytes, _ := strconv.ParseFloat(nw[1], 64)
	rx_packets, _ := strconv.ParseFloat(nw[2], 64)
	tx_bytes, _ := strconv.ParseFloat(nw[9], 64)
	tx_packets, _ := strconv.ParseFloat(nw[10], 64)

	stat = ContainerNetworkStat{
		rx_bytes:   rx_bytes,
		rx_packets: rx_packets,
		tx_bytes:   tx_bytes,
		tx_packets: tx_packets,
	}

	resultList = stat

	return resultList
}
