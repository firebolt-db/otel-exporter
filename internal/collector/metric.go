package collector

import "go.opentelemetry.io/otel/metric"

// runtimeMetrics specifies a set of engine runtime metrics.
type runtimeMetrics struct {
	cpuUtilization    metric.Float64Gauge
	memoryUtilization metric.Float64Gauge
	diskUtilization   metric.Float64Gauge
	cacheUtilization  metric.Float64Gauge
	diskSpilled       metric.Int64UpDownCounter
	runningQueries    metric.Int64Gauge
	suspendedQueries  metric.Int64Gauge
}

// queryHistoryMetrics specifies a set of engine query history metrics.
type queryHistoryMetrics struct {
	queryDuration metric.Float64Histogram

	scannedRows  metric.Int64Counter
	scannedBytes metric.Int64Counter

	insertedRows  metric.Int64Counter
	insertedBytes metric.Int64Counter

	returnedRows  metric.Int64Counter
	returnedBytes metric.Int64Counter
	spilledBytes  metric.Int64Counter

	queueTime            metric.Float64Counter
	queryGatewayDuration metric.Float64Histogram
}

// exporterMetrics specifies a set of supplementary metrics of otel-exporter.
type exporterMetrics struct {
	duration metric.Float64Counter
}

// setupRuntimeMetrics prepares engine runtime metrics with basic attributes and unit.
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
		"firebolt.engine.cache.hit_ratio",
		metric.WithDescription("Current SSD cache hit ratio (percentage)"),
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

	rm.runningQueries, err = meter.Int64Gauge(
		"firebolt.engine.running.queries",
		metric.WithDescription("Number of running queries"),
		metric.WithUnit("{count}"),
	)
	if err != nil {
		return err
	}

	rm.suspendedQueries, err = meter.Int64Gauge(
		"firebolt.engine.suspended.queries",
		metric.WithDescription("Number of suspended queries"),
		metric.WithUnit("{count}"),
	)
	if err != nil {
		return err
	}

	c.runtimeMetrics = rm
	return nil
}

// setupQueryHistoryMetrics prepares engine query history metrics with basic attributes and unit.
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
		"firebolt.query.scanned.bytes",
		metric.WithDescription("The total number of bytes scanned (both from cache and storage)"),
		metric.WithUnit("bytes"),
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
		"firebolt.query.insert.bytes",
		metric.WithDescription("The total number of bytes written (both to cache and storage)"),
		metric.WithUnit("bytes"),
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
		"firebolt.query.returned.bytes",
		metric.WithDescription("The total number of bytes returned from the query"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	qhm.spilledBytes, err = meter.Int64Counter(
		"firebolt.query.spilled.bytes",
		metric.WithDescription("The total number of bytes spilled (uncompressed)"),
		metric.WithUnit("bytes"),
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

	qhm.queryGatewayDuration, err = meter.Float64Histogram(
		"firebolt.query.gateway.duration",
		metric.WithDescription("End to end time the query spent in the gateway"),
		metric.WithUnit("second"),
	)
	if err != nil {
		return err
	}

	c.queryHistoryMetrics = qhm
	return nil
}

// setupExporterMetrics prepares supplementary metrics.
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
