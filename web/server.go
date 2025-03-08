package web

import (
	"html/template"
	"net/http"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/VladMinzatu/go-mon/web/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewServer(systemMonitor *monitor.SystemMonitorService) (http.Handler, error) {
	err := initTelemetry()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	var handler http.Handler = mux
	err = addRoutes(mux, systemMonitor)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

var funcMap = template.FuncMap{
	"toGB": func(bytes uint64) float64 {
		return float64(bytes) / (1024 * 1024 * 1024)
	},
}
var homeTmpl = template.Must(template.New("index.html").ParseFiles("web/views/index.html"))
var statsTmpl = template.Must(template.New("system_monitor.html").Funcs(funcMap).ParseFiles("web/views/system_monitor.html"))

func addRoutes(mux *http.ServeMux, systemMonitor *monitor.SystemMonitorService) error {
	wsHandler, err := handlers.NewWebSocketHandler(systemMonitor, statsTmpl)
	if err != nil {
		return err
	}
	mux.Handle("/", handlers.NewHomepageHandler(homeTmpl))
	mux.Handle("/ws", wsHandler) // test with: websocat ws://localhost:8080/ws
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	mux.Handle("/metrics", promhttp.Handler())
	return nil
}

func initTelemetry() error {
	promExporter, err := prometheus.New()
	if err != nil {
		return err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
	otel.SetMeterProvider(mp)
	return nil
}
