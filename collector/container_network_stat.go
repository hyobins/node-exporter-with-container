package collector

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
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

//GetMemoryStat return container's cpu usage
func getNetworkStat(containers []types.Container) map[string]ContainerNetworkStat {

	resultList := make(map[string]ContainerNetworkStat)
	stat := ContainerNetworkStat{} //initialize

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

		lines, err := readLines("/proc/" + slice[0] + "/net/dev")
		nw := strings.Fields(string(lines[2]))

		rx_bytes, _ := strconv.ParseFloat(nw[1], 64)
		rx_packets, _ := strconv.ParseFloat(nw[2], 64)
		tx_bytes, _ := strconv.ParseFloat(nw[9], 64)
		tx_packets, _ := strconv.ParseFloat(nw[10], 64)
		containerName := strings.Join(container.Names, "")

		stat = ContainerNetworkStat{
			containerID:   container.ID,
			containerName: containerName[1:],
			rx_bytes:      rx_bytes,
			rx_packets:    rx_packets,
			tx_bytes:      tx_bytes,
			tx_packets:    tx_packets,
		}

		resultList[container.ID] = stat
	}

	return resultList
}
