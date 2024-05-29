package cmd

import (
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/firebolt-db/otel-exporter/internal/collector"
	"github.com/firebolt-db/otel-exporter/internal/config"
	"github.com/firebolt-db/otel-exporter/internal/exporter/grpcexporter"
	"github.com/firebolt-db/otel-exporter/internal/exporter/httpexporter"
	"github.com/firebolt-db/otel-exporter/internal/fetcher"
	"github.com/firebolt-db/otel-exporter/internal/logging"
)

type app struct {
	inner *cli.App

	cfg *config.Config
}

func NewApp() *cli.App {
	a := &app{}

	a.inner = &cli.App{
		Name:    "firebolt-otel-exporter",
		Version: "0.1.0",
		Usage:   "The CLI app that starts Firebolt Open Telemetry Exporter.",
		Before:  a.before,
		Action:  a.run,
	}

	return a.inner
}

func (a *app) before(cliCtx *cli.Context) error {
	cfg, err := config.NewConfig(cliCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	logging.Init(cfg.Logging)

	a.cfg = cfg

	return nil
}

func (a *app) run(cliCtx *cli.Context) error {
	ctx := cliCtx.Context
	var err error

	slog.DebugContext(ctx, "starting firebolt opentelemetry exporter")

	f := fetcher.New(a.cfg.Credentials.ClientID, a.cfg.Credentials.ClientSecret)
	slog.DebugContext(ctx, "fetcher initialized")

	var exp metric.Exporter
	if a.cfg.Exporter.GRPC != nil {
		exp, err = grpcexporter.NewGRPCExporter(ctx, a.cfg.Exporter.GRPC)
	} else {
		exp, err = httpexporter.NewHTTPExporter(ctx, a.cfg.Exporter.HTTP)
	}

	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize metrics exporter", slog.Any("error", err))
		return err
	}

	slog.DebugContext(ctx, "exporter initialized")

	col, err := collector.NewCollector(f, a.cfg.Accounts, collector.WithExporter(exp))
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize metrics collector", slog.Any("error", err))
		return err
	}
	slog.InfoContext(ctx, "metrics collector and exporter initialized")

	defer func() {
		if err := col.Close(ctx); err != nil {
			slog.Error("failed to close collector", slog.Any("error", err))
		}
	}()

	return col.Start(ctx, a.cfg.CollectInterval)
}
