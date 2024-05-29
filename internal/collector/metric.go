package collector

import "go.opentelemetry.io/otel/metric"

type runtimeMetrics struct {
	cpuUtilization    metric.Float64Gauge
	memoryUtilization metric.Float64Gauge
	diskUtilization   metric.Float64Gauge
	cacheUtilization  metric.Float64Gauge
	diskSpilled       metric.Int64UpDownCounter
}

type queryHistoryMetrics struct {
	queryDuration metric.Float64Histogram

	scannedRows  metric.Int64Counter
	scannedBytes metric.Int64Counter

	insertedRows  metric.Int64Counter
	insertedBytes metric.Int64Counter

	returnedRows  metric.Int64Counter
	returnedBytes metric.Int64Counter
	spilledBytes  metric.Int64Counter

	queueTime metric.Float64Counter
}

type exporterMetrics struct {
	duration metric.Float64Counter
}

func (c *collector) setupRuntimeMetrics() error {
	meter := c.meterProvider.Meter("firebolt.engine.runtime")

	var err error
	rm := &runtimeMetrics{}

	rm.cpuUtilization, err = meter.Float64Gauge(
		"firebolt.engine.cpu.utilization",
		metric.WithDescription("Current CPU utilization (percentage)"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	rm.memoryUtilization, err = meter.Float64Gauge(
		"firebolt.engine.memory.utilization",
		metric.WithDescription("Current Memory used (percentage)"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	rm.diskUtilization, err = meter.Float64Gauge(
		"firebolt.engine.disk.utilization",
		metric.WithDescription("Currently used disk space which encompasses space used for cache and spilling (percentage)"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	rm.cacheUtilization, err = meter.Float64Gauge(
		"firebolt.engine.cache.utilization",
		metric.WithDescription("Current SSD cache hit ratio"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return err
	}

	rm.diskSpilled, err = meter.Int64UpDownCounter(
		"firebolt.engine.disk.spilled",
		metric.WithDescription("Amount of spilled data to disk in bytes"),
		metric.WithUnit("byte"),
	)
	if err != nil {
		return err
	}

	c.runtimeMetrics = rm
	return nil
}

func (c *collector) setupQueryHistoryMetrics() error {
	meter := c.meterProvider.Meter("firebolt.engine.query_history")

	var err error
	qhm := &queryHistoryMetrics{}

	qhm.queryDuration, err = meter.Float64Histogram(
		"firebolt.query.duration",
		metric.WithDescription("Duration of query execution"),
		metric.WithUnit("second"),
	)
	if err != nil {
		return err
	}

	qhm.scannedRows, err = meter.Int64Counter(
		"firebolt.query.scanned.rows",
		metric.WithDescription("The total number of rows scanned"),
		metric.WithUnit("{row}"),
	)
	if err != nil {
		return err
	}

	qhm.scannedBytes, err = meter.Int64Counter(
		"firebolt.query.scanned.byte",
		metric.WithDescription("The total number of bytes scanned (both from cache and storage)"),
		metric.WithUnit("byte"),
	)
	if err != nil {
		return err
	}

	qhm.insertedRows, err = meter.Int64Counter(
		"firebolt.query.insert.rows",
		metric.WithDescription("The total number of rows written"),
		metric.WithUnit("{row}"),
	)
	if err != nil {
		return err
	}

	qhm.insertedBytes, err = meter.Int64Counter(
		"firebolt.query.insert.byte",
		metric.WithDescription("The total number of bytes written (both to cache and storage)"),
		metric.WithUnit("byte"),
	)
	if err != nil {
		return err
	}

	qhm.returnedRows, err = meter.Int64Counter(
		"firebolt.query.returned.rows",
		metric.WithDescription("The total number of rows returned from the query"),
		metric.WithUnit("{row}"),
	)
	if err != nil {
		return err
	}

	qhm.returnedBytes, err = meter.Int64Counter(
		"firebolt.query.returned.byte",
		metric.WithDescription("The total number of bytes returned from the query"),
		metric.WithUnit("byte"),
	)
	if err != nil {
		return err
	}

	qhm.spilledBytes, err = meter.Int64Counter(
		"firebolt.query.spilled.byte",
		metric.WithDescription("The total number of bytes spilled (uncompressed)"),
		metric.WithUnit("byte"),
	)
	if err != nil {
		return err
	}

	qhm.queueTime, err = meter.Float64Counter(
		"firebolt.query.queue.time",
		metric.WithDescription("Time the query spent in queue"),
		metric.WithUnit("second"),
	)
	if err != nil {
		return err
	}

	c.queryHistoryMetrics = qhm
	return nil
}

func (c *collector) setupExporterMetrics() error {
	meter := c.meterProvider.Meter("firebolt.exporter")

	var err error
	em := &exporterMetrics{}

	em.duration, err = meter.Float64Counter(
		"firebolt.exporter.duration",
		metric.WithDescription("Duration of collection routine of the exporter"),
		metric.WithUnit("second"),
	)
	if err != nil {
		return err
	}

	c.exporterMetrics = em
	return nil
}
