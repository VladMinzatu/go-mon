package telemetry

import (
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	MeterProvider *metric.MeterProvider
	initOnce      sync.Once
	initErr       error

	// Sentinel errors for telemetry initialization and instrument creation failures
	// In a prod setting, we may want to have dedicated alerts for these
	ErrTelemetryInit    = fmt.Errorf("failed to initialize telemetry")
	ErrInstrumentCreate = fmt.Errorf("failed to create telemetry instrument")
)

func InitMetrics() {
	initOnce.Do(func() {
		promExporter, err := prometheus.New()
		if err != nil {
			initErr = ErrTelemetryInit
		}

		mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
		otel.SetMeterProvider(mp)
	})
}

func GetMeter(name string) (metric.Meter, error) {
	InitMetrics() // Ensure it's initialized before use
	if initErr != nil {
		return nil, initErr
	}
	return otel.GetMeterProvider().Meter(name), nil
}
