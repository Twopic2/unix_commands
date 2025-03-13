package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func psContainers(cli *client.Client) ([]types.Container, error) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatal("Error listing containers:", err)
	}
	return containers, nil

}

func psStats(cli *client.Client, containerID string) (float64, float64, error) {
	context := context.Background()
	stats, err := cli.ContainerStats(context, containerID, false)
	if err != nil {
		return 0, 0, err
	}
	defer stats.Body.Close()

	var data types.Stats

	err = json.NewDecoder(stats.Body).Decode(&data)
	if err != nil {
		return 0, 0, err
	}

	cpuDelta := float64(data.CPUStats.CPUUsage.TotalUsage - data.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(data.CPUStats.SystemUsage - data.PreCPUStats.SystemUsage)
	cpuPercent := (cpuDelta / systemDelta) * 100.0

	memUsage := float64(data.MemoryStats.Usage) / (1024 * 1024)
	memLimit := float64(data.MemoryStats.Limit) / (1024 * 1024)
	memPercent := (memUsage / memLimit) * 100.0

	return cpuPercent, memPercent, nil

}

func showContainers(cli *client.Client) {
	for {
		fmt.Println("\nFetching container stats...")
		containers, err := psContainers(cli)
		if err != nil {
			log.Fatal("Error fetching containers:", err)
		}

		for _, container := range containers {
			cpu, mem, err := psStats(cli, container.ID)
			if err != nil {
				log.Printf("Error getting stats for container %s: %v\n", container.ID[:12], err)
				continue
			}

			fmt.Printf("[%s] CPU: %.2f%% | Memory: %.2f%%\n", container.Names[0], cpu, mem)
		}
		time.Sleep(3 * time.Second)
	}

}

func main() {
	stat, err := client.NewClientWithOpts((client.FromEnv))
	if err != nil {
		log.Fatal("Error with Main method", err)
	}

	showContainers(stat)
}
