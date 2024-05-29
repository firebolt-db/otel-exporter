package collector

import (
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
)

type Option interface {
	apply(collector *collector) *collector
}

type optionFunc func(collector *collector) *collector

func (o optionFunc) apply(collector *collector) *collector {
	return o(collector)
}

// WithExporter applies provided exporter to the Collector
func WithExporter(e metric.Exporter) Option {
	return optionFunc(func(collector *collector) *collector {
		collector.exporter = e
		return collector
	})
}

// WithExportInterval applies provided export interval to the Collector
func WithExportInterval(interval time.Duration) Option {
	return optionFunc(func(collector *collector) *collector {
		collector.exportInterval = interval
		return collector
	})
}
