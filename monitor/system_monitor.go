package monitor

import (
	"log/slog"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMetricsProvider interface {
	GetSystemMetrics() (*SystemMetrics, error)
}

type SystemMonitor struct {
	provider SystemMetricsProvider
	interval time.Duration
	metrics  chan *SystemMetrics
	done     chan struct{}
}

func NewSystemMonitor(provider SystemMetricsProvider, interval time.Duration) *SystemMonitor {
	return &SystemMonitor{
		provider: provider,
		interval: interval,
		metrics:  make(chan *SystemMetrics),
		done:     make(chan struct{}),
	}
}

func (m *SystemMonitor) Start() <-chan *SystemMetrics {
	slog.Info("Starting SystemMonitor", slog.String("interval", m.interval.String()), slog.String("message", "Will compute and publish system metrics at regular intervals"))
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		defer close(m.metrics)

		for {
			select {
			case <-m.done:
				slog.Info("SystemMonitor was stopped. Closing the metrics channel.")
				return
			case <-ticker.C:
				metrics, err := m.provider.GetSystemMetrics()
				if err != nil {
					slog.Error("Failed to get system metrics. Cannot publish updates.", "error", err)
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
	cpuPercentages, err := cpu.Percent(time.Second, true) // true means "per CPU core"
	if err != nil {
		// no log - handle errors only once, and propagating is handling
		return nil, err
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		// no log - handle errors only once, and propagating is handling
		return nil, err
	}

	return &SystemMetrics{
		CPUUsagePerCore: cpuPercentages,
		TotalMemory:     memStats.Total,
		UsedMemory:      memStats.Used,
		FreeMemory:      memStats.Available,
		MemoryUsage:     memStats.UsedPercent,
	}, nil
}
