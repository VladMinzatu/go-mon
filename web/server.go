package web

import (
	"html/template"
	"net/http"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/VladMinzatu/go-mon/web/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewServer(systemMonitor *monitor.SystemMonitorService) http.Handler {
	mux := http.NewServeMux()
	var handler http.Handler = mux
	addRoutes(mux, systemMonitor)
	return handler
}

var funcMap = template.FuncMap{
	"toGB": func(bytes uint64) float64 {
		return float64(bytes) / (1024 * 1024 * 1024)
	},
}
var homeTmpl = template.Must(template.New("index.html").ParseFiles("web/views/index.html"))
var statsTmpl = template.Must(template.New("system_monitor.html").Funcs(funcMap).ParseFiles("web/views/system_monitor.html"))

func addRoutes(mux *http.ServeMux, systemMonitor *monitor.SystemMonitorService) {
	mux.Handle("/", handlers.NewHomepageHandler(homeTmpl))
	mux.Handle("/ws", handlers.NewWebSocketHandler(systemMonitor, statsTmpl)) // test with: websocat ws://localhost:8080/ws
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	mux.Handle("/metrics", promhttp.Handler())
}
