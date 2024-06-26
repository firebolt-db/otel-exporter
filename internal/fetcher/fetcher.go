package fetcher

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/firebolt-db/firebolt-go-sdk"
)

// Fetcher is an interface that the metrics fetcher should implement.
type Fetcher interface {
	// FetchEngines reads a list of running engines in a single account.
	FetchEngines(ctx context.Context, accountName string) ([]string, error)

	// FetchRuntimePoints returns a channel of EngineRuntimePoint and pushes data into that channel asynchronously.
	// It should close the channel when all data points are pushed.
	// The metrics should be collected within the provided time interval.
	FetchRuntimePoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan EngineRuntimePoint

	// FetchQueryHistoryPoints returns a channel of QueryHistoryPoint and pushes data into that channel asynchronously
	// It should close the channel when all data points are pushed.
	// The metrics should be collected within the provided time interval.
	FetchQueryHistoryPoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan QueryHistoryPoint
}

// fetcher is an implementation of Fetcher interface.
type fetcher struct {
	clientID, clientSecret string
}

// New creates a new instance of Fetcher, using Firebolt Service Account credentials provided.
func New(clientID, clientSecret string) Fetcher {
	return &fetcher{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// FetchEngines returns a list of running engines in account.
func (f *fetcher) FetchEngines(ctx context.Context, accountName string) ([]string, error) {
	// connect to a system engine to read running engines.
	db, err := f.connect(ctx, accountName, "")
	if err != nil {
		return nil, err
	}

	defer db.Close()

	// we are only interested in running engines.
	rows, err := db.QueryContext(ctx, "SELECT engine_name FROM information_schema.engines WHERE status = 'RUNNING';")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var engines []string

	for rows.Next() {
		var engineName string
		if err := rows.Scan(&engineName); err != nil {
			return nil, err
		}

		engines = append(engines, engineName)
	}

	return engines, nil
}

// FetchRuntimePoints returns a channel of EngineRuntimePoint.
func (f *fetcher) FetchRuntimePoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan EngineRuntimePoint {
	ch := make(chan EngineRuntimePoint)

	go func() {
		wg := sync.WaitGroup{}

		// metrics for each engine are scanned async.
		for _, engine := range engines {
			wg.Add(1)

			go func(engineName string) {
				defer wg.Done()

				// connect to an engine
				engDb, err := f.connect(ctx, account, engineName)
				if err != nil {
					slog.ErrorContext(ctx, "failed to connect to engine",
						slog.String("accountName", account), slog.String("engineName", engineName),
						slog.Any("error", err),
					)
					return
				}
				defer engDb.Close()

				// read the metrics. Only interested in most recent metric within the time interval.
				row := engDb.QueryRowContext(ctx,
					fmt.Sprintf(
						`SELECT engine_cluster, event_time, cpu_used, memory_used, disk_used, 
       						cache_hit_ratio, spilled_bytes, running_queries, suspended_queries  
				FROM information_schema.engine_metrics_history 
         		WHERE event_time > TIMESTAMPTZ '%s' AND event_time <= TIMESTAMPTZ '%s' 
         		ORDER BY event_time DESC LIMIT 1;`,
						since.Format(time.DateTime+"-07"), till.Format(time.DateTime+"-07"),
					))

				// prepare the metric point.
				erp := EngineRuntimePoint{EngineName: engineName}
				if err := erp.Scan(row); err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						slog.ErrorContext(ctx, "failed to scan engine metric",
							slog.String("accountName", account), slog.String("engineName", engineName),
							slog.Any("error", err),
						)
					}
					return
				}

				ch <- erp
			}(engine)
		}

		// wait until all engines metrics are pushed and close the channel
		wg.Wait()
		close(ch)
	}()

	return ch
}

// FetchQueryHistoryPoints returns a channel of QueryHistoryPoint.
func (f *fetcher) FetchQueryHistoryPoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan QueryHistoryPoint {
	ch := make(chan QueryHistoryPoint)

	go func() {
		wg := sync.WaitGroup{}

		// metrics for each engine are scanned async.
		for _, engine := range engines {
			wg.Add(1)

			go func(engineName string) {
				defer wg.Done()

				// connect to an engine
				engDb, err := f.connect(ctx, account, engineName)
				if err != nil {
					slog.ErrorContext(ctx, "failed to connect to engine",
						slog.String("accountName", account), slog.String("engineName", engineName),
						slog.Any("error", err),
					)
					return
				}
				defer engDb.Close()

				// read the metrics within provided time interval. Entries with status='STARTED_EXECUTION' do not provide
				// any metrics data, so they are skipped.
				rows, err := engDb.QueryContext(ctx,
					fmt.Sprintf(
						`SELECT account_name, user_name, duration_us, status, scanned_rows, scanned_bytes, 
       						inserted_rows, inserted_bytes, spilled_bytes, returned_rows, returned_bytes, time_in_queue_us 
					FROM information_schema.engine_query_history
					WHERE status <> 'STARTED_EXECUTION' 
						AND submitted_time > TIMESTAMPTZ '%s' AND submitted_time <= TIMESTAMPTZ '%s' 
         		    ORDER BY submitted_time;`,
						since.Format(time.DateTime+"-07"), till.Format(time.DateTime+"-07"),
					),
				)
				if err != nil {
					slog.ErrorContext(ctx, "failed to read query history metrics",
						slog.String("accountName", account), slog.String("engineName", engineName),
						slog.Any("error", err),
					)
					return
				}

				defer rows.Close()

				// prepare the metric point. There can be multiple queries in history.
				for rows.Next() {
					qhp := QueryHistoryPoint{EngineName: engineName}

					userName, accountName := sql.NullString{}, sql.NullString{}

					if err := rows.Scan(&accountName, &userName, &qhp.DurationMicroSeconds, &qhp.Status,
						&qhp.ScannedRows, &qhp.ScannedBytes, &qhp.InsertedRows, &qhp.InsertedBytes, &qhp.SpilledBytes,
						&qhp.ReturnedRows, &qhp.ReturnedBytes, &qhp.TimeInQueueMicroSeconds,
					); err != nil {
						slog.ErrorContext(ctx, "failed to scan query history metric",
							slog.String("accountName", account), slog.String("engineName", engineName),
							slog.Any("error", err),
						)
						return
					}

					if userName.Valid {
						qhp.UserName = userName.String
					}
					if accountName.Valid {
						qhp.AccountName = accountName.String
					}

					ch <- qhp
				}
			}(engine)
		}

		// wait until all engines metrics are pushed and close the channel
		wg.Wait()
		close(ch)
	}()

	return ch
}

// connect returns a sql.DB instance for specified account and engine. In case engine name is not provided, it will connect
// to a system engine.
func (f *fetcher) connect(ctx context.Context, accountName string, engineName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("firebolt://?account_name=%s&client_id=%s&client_secret=%s",
		accountName, f.clientID, f.clientSecret,
	)

	db, err := sql.Open("firebolt", dsn)
	if err != nil {
		return nil, err
	}

	// switch to the engine if engineName is provided
	if engineName != "" {
		// prevent engine from auto stopping caused by exporter queries
		_, err = db.ExecContext(ctx, `SET auto_start_stop_control=ignore;`)
		if err != nil {
			return nil, fmt.Errorf("failed to set auto_start_stop_control = ignore for engine %s: %w", engineName, err)
		}

		// switch to an engine
		_, err = db.ExecContext(ctx, fmt.Sprintf(`USE ENGINE "%s";`, engineName))
		if err != nil {
			return nil, fmt.Errorf("failed to switch to engine %s: %w", engineName, err)
		}

		// add a query label to appear in query history
		_, err = db.ExecContext(ctx, `SET query_label=otel-exporter;`)
		if err != nil {
			return nil, fmt.Errorf("failed to set query label: %w", err)
		}
	}

	return db, nil
}
