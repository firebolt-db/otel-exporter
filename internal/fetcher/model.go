package fetcher

import (
	"database/sql"
	"time"
)

type EngineRuntimePoint struct {
	EngineName string

	EngineCluster int64
	EventTime     time.Time
	CPUUsed       float64
	MemoryUsed    float64
	DiskUsed      float64
	CacheHitRatio float64
	SpilledBytes  int64
}

func (p *EngineRuntimePoint) Scan(row *sql.Row) error {
	return row.Scan(
		&p.EngineCluster,
		&p.EventTime,
		&p.CPUUsed,
		&p.MemoryUsed,
		&p.DiskUsed,
		&p.CacheHitRatio,
		&p.SpilledBytes,
	)
}

type QueryHistoryPoint struct {
	EngineName string

	AccountName string
	UserName    string

	DurationMicroSeconds int64
	Status               string

	ScannedRows             int64
	ScannedBytes            int64
	InsertedRows            int64
	InsertedBytes           int64
	SpilledBytes            int64
	ReturnedRows            int64
	ReturnedBytes           int64
	TimeInQueueMicroSeconds int64
}
