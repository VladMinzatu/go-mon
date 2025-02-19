package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var funcMap = template.FuncMap{
	"toGB": func(bytes uint64) float64 {
		return float64(bytes) / (1024 * 1024 * 1024)
	},
}

var statsTmpl = template.Must(template.New("system_monitor.html").Funcs(funcMap).ParseFiles("web/views/system_monitor.html"))

type WebSocketHandler struct {
	systemMonitor *monitor.SystemMonitorService
}

func NewWebSocketHandler(systemMonitor *monitor.SystemMonitorService) *WebSocketHandler {
	return &WebSocketHandler{
		systemMonitor: systemMonitor,
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading connection:", "error", err)
		return
	}
	defer ws.Close()

	connClosed := launchConnectionClosedListener(ws)
	metricsChan := h.systemMonitor.Subscribe()
	defer h.systemMonitor.Unsubscribe(metricsChan)
	for {
		select {
		case m := <-metricsChan:
			jsonBytes := toHtml(m)
			if err := ws.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				slog.Error("Error writing message:", "error", err.Error())
				return
			}
		case <-connClosed:
			// Client closed connection
			slog.Debug("Client disconnected")
			return
		}
	}
}

func toHtml(metrics *monitor.SystemMetrics) []byte {
	var buf bytes.Buffer
	if err := statsTmpl.Execute(&buf, metrics); err != nil {
		slog.Error("Error executing template:", "error", err.Error())
		return []byte(err.Error())
	}
	return buf.Bytes()
}

func launchConnectionClosedListener(ws *websocket.Conn) <-chan struct{} {
	connClosed := make(chan struct{})
	// check for client disconnects
	go func() {
		defer close(connClosed)
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.Error("Connection closed unexpectedly:", "error", err)
				}
				return // connection closed in a normal way. No cause for concern and we can now stop trying to send updates
			}
		}
	}()
	return connClosed
}
