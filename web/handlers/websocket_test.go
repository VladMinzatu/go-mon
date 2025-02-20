package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/gorilla/websocket"
)

type mockSystemMonitorService struct {
	metricsChan chan *monitor.SystemMetrics
}

func (m *mockSystemMonitorService) Subscribe() chan *monitor.SystemMetrics {
	return m.metricsChan
}

func (m *mockSystemMonitorService) Unsubscribe(ch chan *monitor.SystemMetrics) {
	close(ch)
}

var testTemplate = template.Must(template.New("index.html").Parse(`
		<!DOCTYPE html>
		<html><body>
		<div id="metrics">
			{{range $index, $usage := .CPUUsagePerCore}}
			<div>Core {{$index}}: {{$usage}}%</div>
			{{end}}
			<div>Memory Usage: {{.MemoryUsage}}%</div>
		</div>
		</body></html>
	`))

func TestWebSocketHandler(t *testing.T) {
	// Create test server
	mockService := &mockSystemMonitorService{
		metricsChan: make(chan *monitor.SystemMetrics),
	}
	handler := NewWebSocketHandler(mockService, testTemplate)
	server := httptest.NewServer(handler)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not open websocket connection: %v", err)
	}
	defer ws.Close()

	testMetrics := &monitor.SystemMetrics{
		CPUUsagePerCore: []float64{50.0},
		TotalMemory:     1024 * 1024 * 1024, // 1GB
		UsedMemory:      512 * 1024 * 1024,  // 512MB
		FreeMemory:      512 * 1024 * 1024,  // 512MB
		MemoryUsage:     50.0,
	}

	go func() {
		mockService.metricsChan <- testMetrics
	}()

	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	expectedStr := "50%"
	if !strings.Contains(string(msg), expectedStr) {
		t.Errorf("expected response to contain CPU usage value '%s', got: %s", expectedStr, string(msg))
	}

	expectedMemStr := "50%"
	if !strings.Contains(string(msg), expectedMemStr) {
		t.Errorf("expected response to contain memory usage value '%s', got: %s", expectedMemStr, string(msg))
	}
}

func TestWebSocketHandlerUpgradeFailure(t *testing.T) {
	mockService := &mockSystemMonitorService{
		metricsChan: make(chan *monitor.SystemMetrics),
	}
	handler := NewWebSocketHandler(mockService, testTemplate)

	// Create a regular HTTP request (not websocket)
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	// This should fail as it's not a proper websocket connection
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
