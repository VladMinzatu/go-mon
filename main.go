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

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jsonBytes := getMetricsJson()
			if err := ws.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				slog.Error("Error writing message:", "error", err.Error())
				return
			}
		}
	}
}

func getMetricsJson() []byte {
	metrics, err := GetSystemMetrics()
	if err != nil {
		slog.Error("Error getting system metrics:", "error", err.Error())
		jsonBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
		return jsonBytes
	}
	jsonBytes, err := json.Marshal(metrics)
	if err != nil {
		slog.Error("Error marshaling metrics:", "error", err.Error())
		jsonBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
		return jsonBytes
	}
	return jsonBytes
}
