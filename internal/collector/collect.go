package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"

	"github.com/firebolt-db/otel-exporter/internal/fetcher"
)

// collectorFn is a function which is responsible for collecting metrics in a single account, with list of engines.
// It is expected that collectorFn will only collect metrics in time interval between `since` and `till`.
type collectorFn func(ctx context.Context, wg *sync.WaitGroup, accountName string, engines []fetcher.Engine, since, till time.Time)

// Start runs main metrics collection routine with specified interval.
// Start will block until provided context is done, or the app is closed.
func (c *collector) Start(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	collectors := []collectorFn{
		c.collectRuntimeMetrics,
		c.collectQueryHistoryMetrics,
	}

	for {
		slog.DebugContext(ctx, "start collecting routine")

		since := c.lastCollectedTime

		collectTime := time.Now().UTC()
		c.lastCollectedTime = collectTime

		func() {
			defer c.reportExporterDuration(ctx, collectTime)

			// run all collectorFns for each account synchronously
			for _, acctName := range c.accounts {
				// fetch engines first, so that the collectorFn doesn't need to
				engines, err := c.fetcher.FetchEngines(ctx, acctName)
				if err != nil {
					slog.Error("failed to fetch engines",
						slog.String("accountName", acctName),
						slog.Any("error", err),
					)
					continue
				}

				wg := &sync.WaitGroup{}
				wg.Add(len(collectors))

				// run all collectors for the account in parallel
				for _, colFn := range collectors {
					go colFn(ctx, wg, acctName, engines, since, collectTime)
				}

				wg.Wait()
			}
		}()

		slog.DebugContext(ctx, "finished collecting routine")

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			continue
		}
	}
}

// reportExporterDuration reports main routine duration counter metric.
func (c *collector) reportExporterDuration(ctx context.Context, startTime time.Time) {
	elapsedSeconds := float64(time.Since(startTime)) / float64(time.Second)
	c.exporterMetrics.duration.Add(ctx, elapsedSeconds)

	slog.DebugContext(ctx, "collecting routine duration", slog.Float64("seconds", elapsedSeconds))
}

// collectRuntimeMetrics collects and reports engine runtime metrics, such as cpu utilization, memory utilization etc.
func (c *collector) collectRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup, accountName string, engines []fetcher.Engine, since, till time.Time) {
	slog.DebugContext(ctx, "start collecting runtime metrics", slog.String("accountName", accountName))

	pointsCh := c.fetcher.FetchRuntimePoints(ctx, accountName, engines, since, till)

	for mp := range pointsCh {
		attrs := []attribute.KeyValue{
			attribute.Key("firebolt.account.name").String(accountName),
			attribute.Key("firebolt.engine.name").String(mp.EngineName),
			attribute.Key("firebolt.engine.status").String(mp.EngineStatus),
		}

		attrsSet := attribute.NewSet(attrs...)

		c.runtimeMetrics.cpuUtilization.Record(ctx, mp.CPUUsed.Float64, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.memoryUtilization.Record(ctx, mp.MemoryUsed.Float64, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.diskUtilization.Record(ctx, mp.DiskUsed.Float64, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.cacheUtilization.Record(ctx, mp.CacheHitRatio.Float64, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.diskSpilled.Add(ctx, mp.SpilledBytes.Int64, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.runningQueries.Record(ctx, mp.RunningQueries.Int64, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.suspendedQueries.Record(ctx, mp.SuspendedQueries.Int64, api.WithAttributeSet(attrsSet))
	}

	wg.Done()

	slog.DebugContext(ctx, "collecting runtime metrics routine finished", slog.String("accountName", accountName))
}

// collectQueryHistoryMetrics collects and reports query history metrics, such as rows and bytes scanned, etc.
func (c *collector) collectQueryHistoryMetrics(ctx context.Context, wg *sync.WaitGroup, accountName string, engines []fetcher.Engine, since, till time.Time) {
	slog.DebugContext(ctx, "start collecting query history metrics", slog.String("accountName", accountName))

	pointsCh := c.fetcher.FetchQueryHistoryPoints(ctx, accountName, engines, since, till)

	for mp := range pointsCh {
		attrs := []attribute.KeyValue{
			attribute.Key("firebolt.account.name").String(accountName),
			attribute.Key("firebolt.engine.name").String(mp.EngineName),
			attribute.Key("firebolt.engine.status").String(mp.EngineStatus),
			attribute.Key("firebolt.user.name").String(mp.UserName.String),
			attribute.Key("firebolt.query.status").String(mp.Status.String),
		}

		attrsSet := attribute.NewSet(attrs...)

		c.queryHistoryMetrics.queryDuration.Record(ctx, float64(mp.DurationMicroSeconds.Int64)/1000000, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.scannedRows.Add(ctx, mp.ScannedRows.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.scannedBytes.Add(ctx, mp.ScannedBytes.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.insertedRows.Add(ctx, mp.InsertedRows.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.insertedBytes.Add(ctx, mp.InsertedBytes.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.returnedRows.Add(ctx, mp.ReturnedRows.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.returnedBytes.Add(ctx, mp.ReturnedBytes.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.spilledBytes.Add(ctx, mp.SpilledBytes.Int64, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.queueTime.Add(ctx, float64(mp.TimeInQueueMicroSeconds.Int64)/1000000, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.queryGatewayDuration.Record(ctx, float64(mp.GatewayDurationMicroSeconds.Int64)/1000000, api.WithAttributeSet(attrsSet))
	}

	wg.Done()

	slog.DebugContext(ctx, "collecting query history metrics routine finished", slog.String("accountName", accountName))
}
