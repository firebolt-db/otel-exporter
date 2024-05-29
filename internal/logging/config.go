package logging

import validation "github.com/go-ozzo/ozzo-validation/v4"

// Config is the logging configuration.
type Config struct {
	// Format specifies the log output format (text or json). By default, json log format is used.
	Format Format `env:"FIREBOLT_OTEL_EXPORTER_LOG_FORMAT,default=json"`

	// Level specifies the log level (debug, info or error). By default, info log level is used.
	Level Level `env:"FIREBOLT_OTEL_EXPORTER_LOG_LEVEL,default=info"`
}

// Validate ensures that Config is valid.
func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(
			&c.Format,
			validation.Required,
			validation.In(FormatText, FormatJSON),
		),
		validation.Field(
			&c.Level,
			validation.Required,
			validation.In(LevelDebug, LevelInfo, LevelError),
		),
	)
}

const (

	// FormatText specifies a text output format.
	FormatText Format = "text"

	// FormatJSON specifies a JSON output format.
	FormatJSON Format = "json"
)

// Format is the log output format.
type Format string

const (
	// LevelDebug specifies the debug log level
	LevelDebug Level = "debug"

	// LevelInfo specifies the info log level.
	LevelInfo Level = "info"

	// LevelError specifies the error log level.
	LevelError Level = "error"
)

// Level is the log level.
type Level string
