package web

import (
	"net/http"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/VladMinzatu/go-mon/web/handlers"
)

func NewServer(systemMonitor *monitor.SystemMonitorService) http.Handler {
	mux := http.NewServeMux()
	var handler http.Handler = mux
	addRoutes(mux, systemMonitor)
	return handler
}

func addRoutes(mux *http.ServeMux, systemMonitor *monitor.SystemMonitorService) {
	mux.Handle("/", &handlers.HomepageHandler{})
	mux.Handle("/ws", handlers.NewWebSocketHandler(systemMonitor)) // test with: websocat ws://localhost:8080/ws
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
}
