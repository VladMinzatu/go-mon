package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/VladMinzatu/go-mon/web/handlers"
)

func main() {
	http.HandleFunc("/", handlers.ServeHomepage)
	http.HandleFunc("/ws", handlers.ServeWs) // test with: websocat ws://localhost:8080/ws
	slog.Info("Starting server on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
