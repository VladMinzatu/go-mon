package handlers

import (
	"log/slog"

	"github.com/VladMinzatu/go-mon/internal/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var (
	// the package-level meter
	meter = telemetry.GetMeter("github.com/VladMinzatu/go-mon/handlers")

	// WebSocket instruments
	numClientsCounter metric.Int64UpDownCounter

	//...more instruments to come
)

func init() {
	var err error

	// init WebSocket instruments
	numClientsCounter, err = meter.Int64UpDownCounter("num_clients",
		metric.WithDescription("Number of active WebSocket connections"),
	)

	if err != nil {
		slog.Error("metrics initialization failed", "error", err)
	}
}
