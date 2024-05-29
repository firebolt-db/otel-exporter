package collector

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/firebolt-db/otel-exporter/internal/fetcher"
)

type Collector interface {
	Close(ctx context.Context) error
	Start(ctx context.Context, interval time.Duration) error
}

type collector struct {
	exporter      metric.Exporter
	meterProvider *metric.MeterProvider
	fetcher       fetcher.Fetcher

	accounts []string

	runtimeMetrics      *runtimeMetrics
	queryHistoryMetrics *queryHistoryMetrics
	exporterMetrics     *exporterMetrics

	lastCollectedTime time.Time

	exportInterval time.Duration
}

func NewCollector(fetcher fetcher.Fetcher, accounts []string, options ...Option) (Collector, error) {
	c := &collector{
		fetcher:           fetcher,
		accounts:          accounts,
		lastCollectedTime: time.Now().UTC(),
		exportInterval:    15 * time.Second,
	}

	for _, opt := range options {
		c = opt.apply(c)
	}

	if c.exporter == nil {
		return nil, fmt.Errorf("must provide either a grpc exporter or a http exporter")
	}

	var err error
	c.meterProvider, err = newMeterProvider(c.exporter, c.exportInterval)
	if err != nil {
		return nil, err
	}

	// set up metrics that will be collected
	if err := c.setupMetrics(); err != nil {
		return nil, err
	}

	// set the meter provider as a global meter provider
	otel.SetMeterProvider(c.meterProvider)

	return c, nil
}

func (c *collector) Close(ctx context.Context) error {
	return c.meterProvider.Shutdown(ctx)
}

func (c *collector) setupMetrics() error {
	if err := c.setupRuntimeMetrics(); err != nil {
		return err
	}

	if err := c.setupQueryHistoryMetrics(); err != nil {
		return err
	}

	if err := c.setupExporterMetrics(); err != nil {
		return err
	}

	return nil
}
