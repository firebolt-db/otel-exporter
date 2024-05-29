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

type Fetcher interface {
	FetchEngines(ctx context.Context, accountName string) ([]string, error)

	FetchRuntimePoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan EngineRuntimePoint
	FetchQueryHistoryPoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan QueryHistoryPoint
}

type fetcher struct {
	clientID, clientSecret string
}

func New(clientID, clientSecret string) Fetcher {
	return &fetcher{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (f *fetcher) FetchEngines(ctx context.Context, accountName string) ([]string, error) {
	db, err := f.connect(ctx, accountName, "")
	if err != nil {
		return nil, err
	}

	defer db.Close()

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

func (f *fetcher) FetchRuntimePoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan EngineRuntimePoint {
	ch := make(chan EngineRuntimePoint)

	go func() {
		wg := sync.WaitGroup{}

		for _, engine := range engines {
			wg.Add(1)

			go func(engineName string) {
				defer wg.Done()

				engDb, err := f.connect(ctx, account, engineName)
				if err != nil {
					slog.ErrorContext(ctx, "failed to connect to engine",
						slog.String("accountName", account), slog.String("engineName", engineName),
						slog.Any("error", err),
					)
					return
				}
				defer engDb.Close()

				// read the metrics
				row := engDb.QueryRowContext(ctx,
					fmt.Sprintf(
						`SELECT engine_cluster, event_time, cpu_used, memory_used, disk_used, cache_hit_ratio, spilled_bytes 
				FROM information_schema.engine_metrics_history 
         		WHERE event_time > TIMESTAMPTZ '%s' AND event_time <= TIMESTAMPTZ '%s' 
         		ORDER BY event_time DESC LIMIT 1;`,
						since.Format(time.DateTime+"-07"), till.Format(time.DateTime+"-07"),
					))

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

		wg.Wait()
		close(ch)
	}()

	return ch
}

func (f *fetcher) FetchQueryHistoryPoints(ctx context.Context, account string, engines []string, since, till time.Time) <-chan QueryHistoryPoint {
	ch := make(chan QueryHistoryPoint)

	go func() {
		wg := sync.WaitGroup{}

		for _, engine := range engines {
			wg.Add(1)

			go func(engineName string) {
				defer wg.Done()

				engDb, err := f.connect(ctx, account, engineName)
				if err != nil {
					slog.ErrorContext(ctx, "failed to connect to engine",
						slog.String("accountName", account), slog.String("engineName", engineName),
						slog.Any("error", err),
					)
					return
				}
				defer engDb.Close()

				// read the metrics
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

				for rows.Next() {
					qhp := QueryHistoryPoint{EngineName: engineName}

					userName := sql.NullString{}

					if err := rows.Scan(&qhp.AccountName, &userName, &qhp.DurationMicroSeconds, &qhp.Status,
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

					ch <- qhp
				}
			}(engine)
		}

		wg.Wait()
		close(ch)
	}()

	return ch
}

func (f *fetcher) connect(ctx context.Context, accountName string, engineName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("firebolt://?account_name=%s&client_id=%s&client_secret=%s",
		accountName, f.clientID, f.clientSecret,
	)

	db, err := sql.Open("firebolt", dsn)
	if err != nil {
		return nil, err
	}

	if engineName != "" {
		_, err = db.ExecContext(ctx, fmt.Sprintf(`USE ENGINE "%s";`, engineName))
		if err != nil {
			return nil, fmt.Errorf("failed to switch to engine %s: %w", engineName, err)
		}
	}

	return db, nil
}
