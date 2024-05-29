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

type Config struct {
	Logging         logging.Config
	Accounts        []string `env:"FIREBOLT_OTEL_EXPORTER_ACCOUNTS"`
	Credentials     Credentials
	Exporter        ExporterConfig
	CollectInterval time.Duration `env:"FIREBOLT_OTEL_EXPORTER_COLLECT_INTERVAL,default=30s"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Logging),
		validation.Field(&c.Accounts, validation.Required),
		validation.Field(&c.Credentials),
		validation.Field(&c.Exporter),
		validation.Field(&c.CollectInterval, validation.Required, validation.Min(15*time.Second)),
	)
}

type Credentials struct {
	ClientID     string `env:"FIREBOLT_OTEL_EXPORTER_CLIENT_ID"`
	ClientSecret string `env:"FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET"`
}

func (c Credentials) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.ClientID, validation.Required),
		validation.Field(&c.ClientSecret, validation.Required),
	)
}

type ExporterConfig struct {
	GRPC *grpcexporter.Config `env:",noinit"`
	HTTP *httpexporter.Config `env:",noinit"`
}

func (c ExporterConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.GRPC,
			validation.When(c.HTTP != nil, validation.Nil),
			validation.When(c.HTTP == nil, validation.NotNil),
		),
		validation.Field(
			&c.HTTP,
			validation.When(c.GRPC != nil, validation.Nil),
			validation.When(c.GRPC == nil, validation.NotNil),
		),
	)
}

func NewConfig(ctx context.Context) (*Config, error) {
	var cfg Config

	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
