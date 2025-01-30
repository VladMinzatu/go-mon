package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading connection:", "error", err)
		return
	}
	defer ws.Close()

	mon := monitor.NewSystemMonitor(1 * time.Second)
	defer mon.Stop()

	connClosed := launchConnectionClosedListener(ws)
	metricsChan := mon.Start(monitor.NewMetricsProvider())
	for {
		select {
		case m := <-metricsChan:
			jsonBytes := toJson(m)
			if err := ws.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				slog.Error("Error writing message:", "error", err.Error())
				return
			}
		case <-connClosed:
			// Client closed connection
			slog.Info("Client disconnected")
			return
		}
	}
}

func toJson(metrics *monitor.SystemMetrics) []byte {
	jsonBytes, err := json.Marshal(*metrics)
	if err != nil {
		slog.Error("Error marshaling metrics:", "error", err.Error())
		jsonBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
		return jsonBytes
	}
	return jsonBytes
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
				return
			}
		}
	}()
	return connClosed
}
