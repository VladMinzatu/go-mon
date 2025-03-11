package telemetry

import (
	"log/slog"
	"os"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	MeterProvider *metric.MeterProvider
	initOnce      sync.Once
)

func InitMetrics() {
	initOnce.Do(func() {
		promExporter, err := prometheus.New()
		if err != nil {
			slog.Error("Failed to initialise server. Shutting down", "error", err.Error())
			os.Exit(1)
		}

		mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
		otel.SetMeterProvider(mp)
	})
}

func GetMeter(name string) metric.Meter {
	InitMetrics() // Ensure it's initialized before use
	return otel.GetMeterProvider().Meter(name)
}
