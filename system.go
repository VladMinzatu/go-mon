package main

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMetrics struct {
	CPUUsagePerCore []float64
	TotalMemory     uint64
	UsedMemory      uint64
	FreeMemory      uint64
	MemoryUsage     float64 // percent
}

func GetSystemMetrics() (*SystemMetrics, error) {
	cpuPercentages, err := cpu.Percent(time.Second, true) // true = per CPU core
	if err != nil {
		return nil, err
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &SystemMetrics{
		CPUUsagePerCore: cpuPercentages,
		TotalMemory:     memStats.Total,
		UsedMemory:      memStats.Used,
		FreeMemory:      memStats.Free,
		MemoryUsage:     memStats.UsedPercent,
	}, nil
}
