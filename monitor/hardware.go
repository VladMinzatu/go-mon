package monitor

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemStats contains information about current system resource usage
type SystemStats struct {
	CPUUsagePerCore []float64 // Percentage usage per CPU core
	TotalMemory     uint64    // Total physical memory in bytes
	UsedMemory      uint64    // Used physical memory in bytes
	FreeMemory      uint64    // Free physical memory in bytes
	MemoryUsage     float64   // Percentage of memory used
}

// GetSystemStats returns current CPU and memory statistics for the system
func GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	// Get CPU usage per core
	cpuPercentages, err := cpu.Percent(0, true) // false = average across all cores, true = per core
	if err != nil {
		return nil, err
	}
	stats.CPUUsagePerCore = cpuPercentages

	// Get memory statistics
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	stats.TotalMemory = memStats.Total
	stats.UsedMemory = memStats.Used
	stats.FreeMemory = memStats.Free
	stats.MemoryUsage = memStats.UsedPercent

	return stats, nil
}
