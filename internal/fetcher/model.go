package fetcher

import (
	"database/sql"
	"time"
)

// Engine represents an engine entry, on which metrics are collected
type Engine struct {
	Name   string
	Status string
}

// EngineRuntimePoint represents a snapshot point of engine runtime metrics.
type EngineRuntimePoint struct {
	EngineName   string
	EngineStatus string

	EngineCluster    sql.NullString
	EventTime        sql.Null[time.Time]
	CPUUsed          sql.NullFloat64
	MemoryUsed       sql.NullFloat64
	DiskUsed         sql.NullFloat64
	CacheHitRatio    sql.NullFloat64
	SpilledBytes     sql.NullInt64
	RunningQueries   sql.NullInt64
	SuspendedQueries sql.NullInt64
}

// Scan fills in EngineRuntimePoint fields from a single row.
func (p *EngineRuntimePoint) Scan(row *sql.Row) error {
	return row.Scan(
		&p.EngineCluster,
		&p.EventTime,
		&p.CPUUsed,
		&p.MemoryUsed,
		&p.DiskUsed,
		&p.CacheHitRatio,
		&p.SpilledBytes,
		&p.RunningQueries,
		&p.SuspendedQueries,
	)
}

// QueryHistoryPoint represents a snapshot point of query history metrics for a single query.
type QueryHistoryPoint struct {
	EngineName   string
	EngineStatus string

	AccountName sql.NullString
	UserName    sql.NullString

	DurationMicroSeconds sql.NullInt64
	Status               sql.NullString

	ScannedRows             sql.NullInt64
	ScannedBytes            sql.NullInt64
	InsertedRows            sql.NullInt64
	InsertedBytes           sql.NullInt64
	SpilledBytes            sql.NullInt64
	ReturnedRows            sql.NullInt64
	ReturnedBytes           sql.NullInt64
	TimeInQueueMicroSeconds sql.NullInt64
}

type TableHistoryPoint struct {
	TableName string

	NumberOfRows      sql.NullInt64
	CompressedBytes   sql.NullInt64
	UncompressedBytes sql.NullInt64
	CompressionRatio  sql.NullFloat64
	NumberOfTablets   sql.NullInt64
	Fragmentation     sql.NullFloat64
}
