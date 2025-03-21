package monitor

import (
	"log/slog"
	"sync"
)

type systemMonitor interface {
	Start() <-chan *SystemMetrics
	Stop()
}

type SystemMonitorService struct {
	monitor       systemMonitor
	subscribers   map[chan *SystemMetrics]struct{}
	mu            sync.RWMutex
	latestMetrics *SystemMetrics
}

func NewSystemMonitorService(monitor systemMonitor) *SystemMonitorService {
	return &SystemMonitorService{
		monitor:     monitor,
		subscribers: make(map[chan *SystemMetrics]struct{}),
	}
}

func (s *SystemMonitorService) Start() {
	metricsChan := s.monitor.Start()

	go func() {
		for metrics := range metricsChan {
			s.latestMetrics = metrics
			s.broadcast(metrics)
		}
	}()
}

func (s *SystemMonitorService) Stop() {
	slog.Info("SystemMonitorService was stopped. Closing all subscriber channels and stopping the monitor.")
	s.mu.Lock()
	defer s.mu.Unlock()

	for ch := range s.subscribers {
		close(ch)
	}
	s.monitor.Stop()
	s.subscribers = make(map[chan *SystemMetrics]struct{})
}

func (s *SystemMonitorService) Subscribe() chan *SystemMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *SystemMetrics, 1) // Buffer of 1 to prevent blocking on write
	if s.latestMetrics != nil {
		ch <- s.latestMetrics // write the cached metrics so the subscriber gets their first update immediately
	}
	s.subscribers[ch] = struct{}{}
	slog.Debug("New subscriber registered")
	return ch
}

func (s *SystemMonitorService) Unsubscribe(ch chan *SystemMetrics) {
	slog.Debug("Unsubscribing a subscriber")
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscribers, ch)
	close(ch)
}

func (s *SystemMonitorService) broadcast(metrics *SystemMetrics) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.subscribers {
		select { // We used buffer size 1 above, but if the channel is still blocked, we'll skip!
		case ch <- metrics:
		default:
		}
	}
}
