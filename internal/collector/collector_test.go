package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	api "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/firebolt-db/otel-exporter/internal/fetcher"
)

func Test_NewCollector(t *testing.T) {
	t.Parallel()

	accounts := []string{"acct1", "acct2"}
	col, err := NewCollector(newFetcherMock(), accounts, WithExporter(newExporterMock()))
	require.NoError(t, err)
	require.NotNil(t, col)

	c := col.(*collector)
	require.Equal(t, accounts, c.accounts)

	require.NotNil(t, c.runtimeMetrics)
	require.NotNil(t, c.runtimeMetrics.cpuUtilization)
	require.NotNil(t, c.runtimeMetrics.memoryUtilization)
	require.NotNil(t, c.runtimeMetrics.diskUtilization)
	require.NotNil(t, c.runtimeMetrics.cacheUtilization)
	require.NotNil(t, c.runtimeMetrics.diskSpilled)

	require.NotNil(t, c.queryHistoryMetrics)
	require.NotNil(t, c.queryHistoryMetrics.queryDuration)
	require.NotNil(t, c.queryHistoryMetrics.scannedRows)
	require.NotNil(t, c.queryHistoryMetrics.scannedBytes)
	require.NotNil(t, c.queryHistoryMetrics.insertedRows)
	require.NotNil(t, c.queryHistoryMetrics.insertedBytes)
	require.NotNil(t, c.queryHistoryMetrics.returnedRows)
	require.NotNil(t, c.queryHistoryMetrics.returnedBytes)
	require.NotNil(t, c.queryHistoryMetrics.spilledBytes)
	require.NotNil(t, c.queryHistoryMetrics.queueTime)
	require.NotNil(t, c.queryHistoryMetrics.gatewayTime)

	require.NotNil(t, c.exporterMetrics)
	require.NotNil(t, c.exporterMetrics.duration)

	require.NoError(t, col.Close(context.Background()))
}

func Test_NewCollector_missing_exporter(t *testing.T) {
	t.Parallel()

	col, err := NewCollector(newFetcherMock(), []string{"acc"})
	require.ErrorContains(t, err, "must provide either a grpc exporter or a http exporter")
	require.Nil(t, col)
}

type fetcherMock struct {
	fetchEnginesFn            func(ctx context.Context, accountName string) ([]fetcher.Engine, error)
	fetchRuntimePointsFn      func(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.EngineRuntimePoint
	fetchQueryHistoryPointsFn func(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.QueryHistoryPoint
}

func newFetcherMock() *fetcherMock {
	return &fetcherMock{
		fetchEnginesFn: func(ctx context.Context, accountName string) ([]fetcher.Engine, error) {
			panic("default FetchEngines")
		},
		fetchRuntimePointsFn: func(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.EngineRuntimePoint {
			panic("default FetchRuntimePoints")
		},
		fetchQueryHistoryPointsFn: func(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.QueryHistoryPoint {
			panic("default FetchQueryHistoryPoints")
		},
	}
}

func (m *fetcherMock) FetchEngines(ctx context.Context, accountName string) ([]fetcher.Engine, error) {
	return m.fetchEnginesFn(ctx, accountName)
}
func (m *fetcherMock) FetchRuntimePoints(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.EngineRuntimePoint {
	return m.fetchRuntimePointsFn(ctx, account, engines, since, till)
}
func (m *fetcherMock) FetchQueryHistoryPoints(ctx context.Context, account string, engines []fetcher.Engine, since, till time.Time) <-chan fetcher.QueryHistoryPoint {
	return m.fetchQueryHistoryPointsFn(ctx, account, engines, since, till)
}

type exporterMock struct {
	temporalityFn api.TemporalitySelector
	aggregationFn api.AggregationSelector
	exportFn      func(context.Context, *metricdata.ResourceMetrics) error
	flushFn       func(context.Context) error
	shutdownFn    func(context.Context) error
}

var _ api.Exporter = (*exporterMock)(nil)

func newExporterMock() *exporterMock {
	return &exporterMock{}
}

func (e *exporterMock) Temporality(k api.InstrumentKind) metricdata.Temporality {
	if e.temporalityFn != nil {
		return e.temporalityFn(k)
	}
	return api.DefaultTemporalitySelector(k)
}

func (e *exporterMock) Aggregation(k api.InstrumentKind) api.Aggregation {
	if e.aggregationFn != nil {
		return e.aggregationFn(k)
	}
	return api.DefaultAggregationSelector(k)
}

func (e *exporterMock) Export(ctx context.Context, m *metricdata.ResourceMetrics) error {
	if e.exportFn != nil {
		return e.exportFn(ctx, m)
	}
	return nil
}

func (e *exporterMock) ForceFlush(ctx context.Context) error {
	if e.flushFn != nil {
		return e.flushFn(ctx)
	}
	return nil
}

func (e *exporterMock) Shutdown(ctx context.Context) error {
	if e.shutdownFn != nil {
		return e.shutdownFn(ctx)
	}
	return nil
}
