package main

import (
	"github.com/patwie/cluster-smi/cluster"
	"github.com/patwie/cluster-smi/nvml"
	"github.com/patwie/cluster-smi/proc"
	"os"
	"os/user"
	"strconv"
	"time"
)

// Cluster
func FetchCluster(c *cluster.Cluster) {
	for i, _ := range c.Nodes {
		FetchNode(&c.Nodes[i])
	}
}

// Node
func InitNode(n *cluster.Node) {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	n.Name = name
	n.Time = time.Now()

	devices, _ := nvml.GetDevices()

	for i := 0; i < len(devices); i++ {
		n.Devices = append(n.Devices, cluster.Device{0, "", 0, cluster.Memory{0, 0, 0, 0}, nil})
	}
}

func FetchNode(n *cluster.Node) {

	devices, _ := nvml.GetDevices()
	n.Time = time.Now()

	for idx, device := range devices {

		meminfo, _ := device.GetMemoryInfo()
		gpuPercent, _, _ := device.GetUtilization()
		memPercent := int(meminfo.Used / meminfo.Total)

		// read processes
		deviceProcs, err := device.GetProcessInfo()
		if err != nil {
			panic(err)
		}

		// collect al proccess informations
		var processes []cluster.Process

		for i := 0; i < len(deviceProcs); i++ {

			if int(deviceProcs[i].Pid) == 0 {
				continue
			}

			PID := deviceProcs[i].Pid
			_, name := proc.TimeAndNameFromPID(PID)

			UID := proc.UIDFromPID(PID)
			user, err := user.LookupId(strconv.Itoa(UID))

			username := "unknown"
			if err == nil {
				username = user.Username
			}

			processes = append(processes, cluster.Process{
				Pid:           PID,
				UsedGpuMemory: deviceProcs[i].UsedGpuMemory,
				Name:          name,
				Username:      username,
			})
		}

		n.Devices[idx].Id = idx
		n.Devices[idx].Name = device.DeviceName
		n.Devices[idx].Utilization = gpuPercent
		n.Devices[idx].MemoryUtilization = cluster.Memory{meminfo.Used, meminfo.Free, meminfo.Total, memPercent}
		n.Devices[idx].Processes = processes

	}
}
