package monitor

import (
	"sync"
	"testing"
	"time"
)

// we'll use a mocked SystemMonitor instance in our tests
type mockSystemMonitor struct {
	metricsChan chan *SystemMetrics
	stopCalled  bool
	mu          sync.Mutex
}

func newMockSystemMonitor() *mockSystemMonitor {
	return &mockSystemMonitor{
		metricsChan: make(chan *SystemMetrics),
	}
}

func (m *mockSystemMonitor) Start() <-chan *SystemMetrics {
	return m.metricsChan
}

func (m *mockSystemMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalled = true
}

var mockMetrics = &SystemMetrics{
	CPUUsagePerCore: []float64{50.0},
	TotalMemory:     1000,
	UsedMemory:      500,
	FreeMemory:      500,
	MemoryUsage:     50.0,
}

func TestSystemMonitorService(t *testing.T) {
	t.Run("subscribers receive metrics", func(t *testing.T) {
		mock := newMockSystemMonitor()
		service := NewSystemMonitorService(mock)
		service.Start()

		// we have 2 new subscribers
		ch1 := service.Subscribe()
		ch2 := service.Subscribe()

		mock.metricsChan <- mockMetrics

		// Both subscribers should receive the metrics
		select {
		case received := <-ch1:
			if received != mockMetrics {
				t.Error("ch1 received incorrect metrics")
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for ch1")
		}

		select {
		case received := <-ch2:
			if received != mockMetrics {
				t.Error("ch2 received incorrect metrics")
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for ch2")
		}
	})

	t.Run("unsubscribe removes channel", func(t *testing.T) {
		mock := newMockSystemMonitor()
		service := NewSystemMonitorService(mock)

		ch := service.Subscribe()
		if len(service.subscribers) != 1 {
			t.Error("expected 1 subscriber")
		}

		service.Unsubscribe(ch)
		if len(service.subscribers) != 0 {
			t.Error("expected 0 subscribers")
		}

		if _, ok := <-ch; ok {
			t.Error("channel should be closed")
		}
	})

	t.Run("calling Stop() closes all subscribers", func(t *testing.T) {
		mock := newMockSystemMonitor()
		service := NewSystemMonitorService(mock)

		ch1 := service.Subscribe()
		ch2 := service.Subscribe()

		service.Stop()

		if _, ok := <-ch1; ok {
			t.Error("ch1 should be closed")
		}
		if _, ok := <-ch2; ok {
			t.Error("ch2 should be closed")
		}

		mock.mu.Lock()
		if !mock.stopCalled {
			t.Error("underlying monitor Stop() was not called")
		}
		mock.mu.Unlock()
	})

	t.Run("new subscriber gets cached metrics immediately upon subscribing", func(t *testing.T) {
		mock := newMockSystemMonitor()
		service := NewSystemMonitorService(mock)
		service.Start()

		ch1 := service.Subscribe()
		// Start a goroutine to receive from ch1 and verify it blocks
		received := make(chan struct{})
		go func() {
			<-ch1 // This should block until we send metrics
			close(received)
		}()

		// Verify channel is blocked
		select {
		case <-received:
			t.Error("channel should block until metrics are sent")
		case <-time.After(100 * time.Millisecond):
		}

		// Now send metrics and verify it unblocks
		mock.metricsChan <- mockMetrics

		select {
		case <-received:
			// channel received metrics and unblocked, all good
		case <-time.After(100 * time.Millisecond):
			t.Error("channel should have received metrics and unblocked")
		}

		ch2 := service.Subscribe()
		// Verify ch2 gets cached metrics immediately
		select {
		case metrics := <-ch2:
			if metrics != mockMetrics {
				t.Error("new subscriber should receive cached metrics immediately")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("new subscriber should receive cached metrics without delay")
		}
	})
}
