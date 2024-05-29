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

func WithExporter(e metric.Exporter) Option {
	return optionFunc(func(collector *collector) *collector {
		collector.exporter = e
		return collector
	})
}

func WithExportInterval(interval time.Duration) Option {
	return optionFunc(func(collector *collector) *collector {
		collector.exportInterval = interval
		return collector
	})
}
