package collector

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/firebolt-db/otel-exporter/internal/fetcher"
)

// Collector defines an interface that the collector should implement.
type Collector interface {
	// Close should be called to clean up allocated resources
	Close(ctx context.Context) error
	// Start is a blocking function which should run the main collector's process
	Start(ctx context.Context, interval time.Duration) error
}

// collector is an implementation of Collector interface.
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

// NewCollector creates a new instance of the [Collector] that will observe a list of accounts.
func NewCollector(fetcher fetcher.Fetcher, accounts []string, options ...Option) (Collector, error) {
	c := &collector{
		fetcher:           fetcher,
		accounts:          accounts,
		lastCollectedTime: time.Now().UTC(), // start observing metrics from current timestamp.
		exportInterval:    15 * time.Second, // default export interval, which defines how often metrics will be pushed to collector.
	}

	for _, opt := range options {
		c = opt.apply(c)
	}

	// check that exporter option was applied.
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

// Close will close allocated meter provider.
func (c *collector) Close(ctx context.Context) error {
	return c.meterProvider.Shutdown(ctx)
}

// setupMetrics prepares all the metrics reported by the collector.
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
