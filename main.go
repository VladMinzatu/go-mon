package main

import (
	"encoding/json"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const pingPeriod = 1 * time.Second

func main() {
	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type data struct {
			Heading string
		}
		tmpl.Execute(w, data{Heading: "Heading is templated"})
	})
	http.HandleFunc("/ws", serveWs) // test with: websocat ws://localhost:8080/ws
	slog.Info("Starting server on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading connection:", "error", err)
		return
	}
	defer ws.Close()

	mon := NewSystemMonitor(1 * time.Second)
	defer mon.Stop()

	connClosed := launchConnectionClosedListener(ws)
	metricsChan := mon.Start(NewMetricsProvider())
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

func toJson(metrics *SystemMetrics) []byte {
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
					slog.Error("Read error:", "error", err)
				}
				return
			}
		}
	}()
	return connClosed
}
