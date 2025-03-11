package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/gorilla/websocket"
)

const (
	numClientsCounterName = "num_clients"
)

type systemMonitorService interface {
	Subscribe() chan *monitor.SystemMetrics
	Unsubscribe(chan *monitor.SystemMetrics)
}

type WebSocketHandler struct {
	upgrader      websocket.Upgrader
	systemMonitor systemMonitorService
	template      *template.Template
}

func NewWebSocketHandler(systemMonitor systemMonitorService, tmpl *template.Template) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		systemMonitor: systemMonitor,
		template:      tmpl,
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	numClientsCounter.Add(r.Context(), 1)
	ws, err := h.upgrader.Upgrade(w, r, nil)
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
			jsonBytes := toHtml(m, h.template)
			if err := ws.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				slog.Error("Error writing message:", "error", err.Error())
				numClientsCounter.Add(r.Context(), -1)
				return
			}
		case <-connClosed:
			// Client closed connection
			slog.Debug("Client disconnected")
			numClientsCounter.Add(r.Context(), -1)
			return
		}
	}
}

func toHtml(metrics *monitor.SystemMetrics, tmpl *template.Template) []byte {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, metrics); err != nil {
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
