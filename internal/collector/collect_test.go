package collector

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/firebolt-db/otel-exporter/internal/fetcher"
)

func Test_Collector_Start(t *testing.T) {
	t.Parallel()

	acctName := "acct"
	databaseName := "test_db"
	f := newFetcherMock()
	exp := newExporterMock()
	c, err := NewCollector(f, []string{acctName}, WithExporter(exp), WithExportInterval(30*time.Millisecond))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	t.Cleanup(cancel)

	interval := 25 * time.Millisecond

	eng := []fetcher.Engine{
		{Name: "engine1", Status: "RUNNING"},
		{Name: "engine2", Status: "RESIZING"},
	}
	f.fetchEnginesFn = func(ctx context.Context, accountName string) ([]fetcher.Engine, error) {
		require.Equal(t, acctName, accountName)
		return eng, nil
	}
	rCh := make(chan fetcher.EngineRuntimePoint)
	f.fetchRuntimePointsFn = func(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.EngineRuntimePoint {
		require.Equal(t, acctName, account)
		require.Equal(t, eng, engines)
		return rCh
	}
	qhCh := make(chan fetcher.QueryHistoryPoint)
	f.fetchQueryHistoryPointsFn = func(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.QueryHistoryPoint {
		require.Equal(t, acctName, account)
		require.Equal(t, eng, engines)
		return qhCh
	}

	thCh := make(chan fetcher.TableHistoryPoint)
	f.fetchTableHistoryPointsFn = func(ctx context.Context, account string, engines []fetcher.Engine, database string) <-chan fetcher.TableHistoryPoint {
		require.Equal(t, acctName, account)
		require.Equal(t, eng, engines)
		require.Equal(t, databaseName, database)
		return thCh
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
			EngineName:   "eng1",
			EngineStatus: "RUNNING",
			EventTime:    sql.Null[time.Time]{Valid: true, V: time.Now()},
			CPUUsed:      sql.NullFloat64{Valid: true, Float64: 10},
		}

		qhCh <- fetcher.QueryHistoryPoint{
			EngineName:           "eng2",
			EngineStatus:         "RESIZING",
			DurationMicroSeconds: sql.NullInt64{Valid: true, Int64: 10},
		}
		thCh <- fetcher.TableHistoryPoint{
			TableName:         "test",
			NumberOfRows:      sql.NullInt64{Int64: 1},
			CompressedBytes:   sql.NullInt64{Int64: 1},
			UncompressedBytes: sql.NullInt64{Int64: 2},
			CompressionRatio:  sql.NullFloat64{Float64: 0.5},
			NumberOfTablets:   sql.NullInt64{Int64: 1},
			Fragmentation:     sql.NullFloat64{Float64: 0.0},
		}
		sentCh <- struct{}{}
	}()

	doneCh := make(chan struct{})
	go func() {
		err := c.Start(ctx, interval, databaseName)
		require.NoError(t, err)
		doneCh <- struct{}{}
	}()

	<-sentCh
	close(rCh)
	close(qhCh)
	close(thCh)
	<-doneCh

	require.Eventually(t, func() bool {
		return exportCalled.Load()
	}, 1000*time.Millisecond, 10*time.Millisecond)
}
