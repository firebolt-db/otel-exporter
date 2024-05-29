package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

type collectorFn func(ctx context.Context, wg *sync.WaitGroup, accountName string, engines []string, since, till time.Time)

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

			for _, acctName := range c.accounts {
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

func (c *collector) reportExporterDuration(ctx context.Context, startTime time.Time) {
	elapsedSeconds := float64(time.Since(startTime)) / float64(time.Second)
	c.exporterMetrics.duration.Add(ctx, elapsedSeconds)

	slog.DebugContext(ctx, "collecting routine duration", slog.Float64("seconds", elapsedSeconds))
}

func (c *collector) collectRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup, accountName string, engines []string, since, till time.Time) {
	slog.DebugContext(ctx, "start collecting runtime metrics", slog.String("accountName", accountName))

	pointsCh := c.fetcher.FetchRuntimePoints(ctx, accountName, engines, since, till)

	for mp := range pointsCh {
		attrs := []attribute.KeyValue{
			attribute.Key("firebolt.account.name").String(accountName),
			attribute.Key("firebolt.engine.name").String(mp.EngineName),
		}

		attrsSet := attribute.NewSet(attrs...)

		c.runtimeMetrics.cpuUtilization.Record(ctx, mp.CPUUsed, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.memoryUtilization.Record(ctx, mp.MemoryUsed, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.diskUtilization.Record(ctx, mp.DiskUsed, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.cacheUtilization.Record(ctx, mp.CacheHitRatio, api.WithAttributeSet(attrsSet))
		c.runtimeMetrics.diskSpilled.Add(ctx, mp.SpilledBytes, api.WithAttributeSet(attrsSet))
	}

	wg.Done()

	slog.DebugContext(ctx, "collecting runtime metrics routine finished", slog.String("accountName", accountName))
}

func (c *collector) collectQueryHistoryMetrics(ctx context.Context, wg *sync.WaitGroup, accountName string, engines []string, since, till time.Time) {
	slog.DebugContext(ctx, "start collecting query history metrics", slog.String("accountName", accountName))

	pointsCh := c.fetcher.FetchQueryHistoryPoints(ctx, accountName, engines, since, till)

	for mp := range pointsCh {
		attrs := []attribute.KeyValue{
			attribute.Key("firebolt.account.name").String(accountName),
			attribute.Key("firebolt.engine.name").String(mp.EngineName),
			attribute.Key("firebolt.user.name").String(mp.UserName),
			attribute.Key("firebolt.query.status").String(mp.Status),
		}

		attrsSet := attribute.NewSet(attrs...)

		c.queryHistoryMetrics.queryDuration.Record(ctx, float64(mp.DurationMicroSeconds)/1000000, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.scannedRows.Add(ctx, mp.ScannedRows, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.scannedBytes.Add(ctx, mp.ScannedBytes, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.insertedRows.Add(ctx, mp.InsertedRows, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.insertedBytes.Add(ctx, mp.InsertedBytes, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.returnedRows.Add(ctx, mp.ReturnedRows, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.returnedBytes.Add(ctx, mp.ReturnedBytes, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.spilledBytes.Add(ctx, mp.SpilledBytes, api.WithAttributeSet(attrsSet))
		c.queryHistoryMetrics.queueTime.Add(ctx, float64(mp.TimeInQueueMicroSeconds)/1000000, api.WithAttributeSet(attrsSet))
	}

	wg.Done()

	slog.DebugContext(ctx, "collecting query history metrics routine finished", slog.String("accountName", accountName))
}
