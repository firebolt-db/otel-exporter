package logging

import (
	"log/slog"
	"os"
)

// Init prepares logger for usage. It is assumed that log/slog package is used.
func Init(cfg Config) {
	var handlerOpts *slog.HandlerOptions

	// Prepare log level. Default level is slog.LevelInfo.
	switch cfg.Level {
	case LevelDebug:
		handlerOpts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	case LevelError:
		handlerOpts = &slog.HandlerOptions{
			Level: slog.LevelError,
		}
	default:
		handlerOpts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	var h slog.Handler

	// Prepare log format. Default log format is json.
	switch cfg.Format {
	case FormatText:
		h = slog.NewTextHandler(os.Stderr, handlerOpts)
	default:
		h = slog.NewJSONHandler(os.Stderr, handlerOpts)
	}

	// apply logging settings to default log/slog logger.
	slog.SetDefault(slog.New(h))
}
