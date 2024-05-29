package logging

import (
	"log/slog"
	"os"
)

func Init(cfg Config) {
	var handlerOpts *slog.HandlerOptions
	// default level is slog.LevelInfo
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
	// default format is json
	switch cfg.Format {
	case FormatText:
		h = slog.NewTextHandler(os.Stderr, handlerOpts)
	default:
		h = slog.NewJSONHandler(os.Stderr, handlerOpts)
	}

	slog.SetDefault(slog.New(h))
}
