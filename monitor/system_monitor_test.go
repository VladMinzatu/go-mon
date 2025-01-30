package monitor

import (
	"testing"
	"time"
)

var monitorPeriod = 10 * time.Millisecond

type testMetricsProvider struct{}

func (tp *testMetricsProvider) GetSystemMetrics() (*SystemMetrics, error) {
	return &SystemMetrics{
		CPUUsagePerCore: []float64{1.0, 2.0},
		TotalMemory:     1000,
		UsedMemory:      500,
		FreeMemory:      500,
		MemoryUsage:     50.0,
	}, nil
}

func TestSystemMonitor(t *testing.T) {
	tests := []struct {
		name          string
		intervalCount int
	}{
		{
			name:          "receives expected number of metrics",
			intervalCount: 3,
		},
		{
			name:          "stops after requested intervals",
			intervalCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewSystemMonitor(monitorPeriod)
			metricsChan := monitor.Start(&testMetricsProvider{})

			receivedCount := 0
			timeout := time.After(monitorPeriod * time.Duration(tt.intervalCount+1))

			for receivedCount < tt.intervalCount {
				select {
				case metric := <-metricsChan:
					if metric == nil {
						t.Error("received nil metrics")
					}
					receivedCount++
				case <-timeout:
					t.Errorf("timed out waiting for metrics, received only %d of %d", receivedCount, tt.intervalCount)
					return
				}
			}

			monitor.Stop()

			// Verify channel is closed after Stop()
			select {
			case _, ok := <-metricsChan:
				if ok {
					t.Error("metrics channel not closed after Stop()")
				}
			case <-time.After(monitorPeriod):
				t.Error("timeout waiting for channel to close")
			}
		})
	}
}
