package collector

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"

	"github.com/firebolt-db/otel-exporter/internal/fetcher"
)

func Test_Collector_Start(t *testing.T) {
	t.Parallel()

	acctName := "acct"
	f := newFetcherMock()
	exp := newExporterMock()
	c, err := NewCollector(f, []string{acctName}, WithExporter(exp), WithExportInterval(30*time.Millisecond))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	t.Cleanup(cancel)

	interval := 25 * time.Millisecond

	eng := []string{"engine1", "engine2"}
	f.fetchEnginesFn = func(ctx context.Context, accountName string) ([]string, error) {
		require.Equal(t, acctName, accountName)
		return eng, nil
	}
	rCh := make(chan fetcher.EngineRuntimePoint)
	f.fetchRuntimePointsFn = func(ctx context.Context, account string, engines []string, since, till time.Time) <-chan fetcher.EngineRuntimePoint {
		require.Equal(t, acctName, account)
		require.Equal(t, eng, engines)
		return rCh
	}
	qhCh := make(chan fetcher.QueryHistoryPoint)
	f.fetchQueryHistoryPointsFn = func(ctx context.Context, account string, engines []string, since, till time.Time) <-chan fetcher.QueryHistoryPoint {
		require.Equal(t, acctName, account)
		require.Equal(t, eng, engines)
		return qhCh
	}
	exportCalled := atomic.Bool{}
	exportCalled.Store(false)
	exp.exportFn = func(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
		require.NotNil(t, metrics)
		require.Equal(t, semconv.SchemaURL, metrics.Resource.SchemaURL())
		attrs := metrics.Resource.Attributes()
		require.Contains(t, attrs, attribute.Key("service.name").String(serviceName))
		require.Contains(t, attrs, attribute.Key("service.version").String(Version))
		require.NotEmpty(t, metrics.ScopeMetrics)

		exportCalled.Store(true)
		return nil
	}

	// push some data as metrics
	sentCh := make(chan struct{})
	go func() {
		rCh <- fetcher.EngineRuntimePoint{
			EngineName: "eng1",
			EventTime:  time.Now(),
			CPUUsed:    10,
		}

		qhCh <- fetcher.QueryHistoryPoint{
			EngineName:           "eng2",
			DurationMicroSeconds: 10,
		}
		sentCh <- struct{}{}
	}()

	doneCh := make(chan struct{})
	go func() {
		err := c.Start(ctx, interval)
		require.NoError(t, err)
		doneCh <- struct{}{}
	}()

	<-sentCh
	close(rCh)
	close(qhCh)
	<-doneCh

	require.Eventually(t, func() bool {
		return exportCalled.Load()
	}, 1000*time.Millisecond, 10*time.Millisecond)
}
