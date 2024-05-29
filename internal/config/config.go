package config

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sethvargo/go-envconfig"

	"github.com/firebolt-db/otel-exporter/internal/exporter/grpcexporter"
	"github.com/firebolt-db/otel-exporter/internal/exporter/httpexporter"
	"github.com/firebolt-db/otel-exporter/internal/logging"
)

// Config defines app configuration. It is expected that all the values in configuration are provided via
// environment variables.
type Config struct {
	// Logging specifies configuration of the logging.
	Logging logging.Config

	// Accounts specifies a list of the accounts, which will be observed by the collector.
	// At least one account must be provided.
	Accounts []string `env:"FIREBOLT_OTEL_EXPORTER_ACCOUNTS"`

	// Credentials specifies Firebolt Service Account credentials, used to run queries.
	Credentials Credentials

	// Exporter specifies configuration of the exporter. Only one of GRPC or HTTP exporter is allowed.
	Exporter ExporterConfig

	// CollectInterval specifies how often otel-exporter will collect metrics from Firebolt. It will also define
	// a discretion step of reported metrics.
	CollectInterval time.Duration `env:"FIREBOLT_OTEL_EXPORTER_COLLECT_INTERVAL,default=30s"`
}

// Validate validates Config
func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Logging),
		validation.Field(&c.Accounts, validation.Required),
		validation.Field(&c.Credentials),
		validation.Field(&c.Exporter),
		// Minimal allowed collect interval is 15s.
		validation.Field(&c.CollectInterval, validation.Required, validation.Min(15*time.Second)),
	)
}

// Credentials specifies Firebolt Service Account credentials, used to run queries.
type Credentials struct {
	// ClientID is client_id of the Firebolt Service Account
	ClientID string `env:"FIREBOLT_OTEL_EXPORTER_CLIENT_ID"`

	// ClientSecret is client_secret of the Firebolt Service Account
	ClientSecret string `env:"FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET"`
}

// Validate validates Credentials.
func (c Credentials) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.ClientID, validation.Required),
		validation.Field(&c.ClientSecret, validation.Required),
	)
}

// ExporterConfig specifies configuration of the exporter.
type ExporterConfig struct {
	// GRPC specifies grpc exporter configuration
	GRPC *grpcexporter.Config `env:",noinit"`

	// HTTP specifies http exporter configuration.
	HTTP *httpexporter.Config `env:",noinit"`
}

// Validate validates ExporterConfig.
func (c ExporterConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.GRPC, // GRPC config must be provided when HTTP config is not set, and not allowed otherwise
			validation.When(c.HTTP != nil, validation.Nil),
			validation.When(c.HTTP == nil, validation.NotNil),
		),
		validation.Field(
			&c.HTTP, // HTTP config must be provided when GRPC config is not set, and not allowed otherwise
			validation.When(c.GRPC != nil, validation.Nil),
			validation.When(c.GRPC == nil, validation.NotNil),
		),
	)
}

// NewConfig creates a new instance of Config. It is expected that all configuration variables are passed
// via environment. Returns error in case config can't be parsed or is invalid.
func NewConfig(ctx context.Context) (*Config, error) {
	var cfg Config

	// process the configuration from the env variables
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}

	// validate config before passing it to other components.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
