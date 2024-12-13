package collector

import (
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	serviceName = "firebolt-otel-exporter"
)

var Version = "v0.0.0-dev"

// newMeterProvider create a new opentelemetry meter provider, and instruments it with basic resource.
func newMeterProvider(exporter metric.Exporter, interval time.Duration) (*metric.MeterProvider, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(Version),
		),
	)
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(exporter,
				metric.WithInterval(interval),
			),
		),
	)

	return mp, nil
}
