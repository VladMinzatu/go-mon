package main

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMonitor struct {
	interval time.Duration
	metrics  chan *SystemMetrics
	done     chan struct{}
}

func NewSystemMonitor(interval time.Duration) *SystemMonitor {
	return &SystemMonitor{
		interval: interval,
		metrics:  make(chan *SystemMetrics),
		done:     make(chan struct{}),
	}
}

type SystemMetricsProvider interface {
	GetSystemMetrics() (*SystemMetrics, error)
}

func (m *SystemMonitor) Start(provider SystemMetricsProvider) <-chan *SystemMetrics {
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		defer close(m.metrics)

		for {
			select {
			case <-m.done:
				return
			case <-ticker.C:
				metrics, err := provider.GetSystemMetrics()
				if err != nil {
					continue // Skip this interval if there's an error
				}
				m.metrics <- metrics
			}
		}
	}()
	return m.metrics
}

func (m *SystemMonitor) Stop() {
	close(m.done)
}

type DefaultSystemMetricsProvider struct {
}

func NewMetricsProvider() *DefaultSystemMetricsProvider {
	return &DefaultSystemMetricsProvider{}
}

type SystemMetrics struct {
	CPUUsagePerCore []float64
	TotalMemory     uint64
	UsedMemory      uint64
	FreeMemory      uint64
	MemoryUsage     float64 // percent
}

func (mp *DefaultSystemMetricsProvider) GetSystemMetrics() (*SystemMetrics, error) {
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
