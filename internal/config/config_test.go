package config_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/firebolt-db/otel-exporter/internal/config"
	"github.com/firebolt-db/otel-exporter/internal/exporter/grpcexporter"
	"github.com/firebolt-db/otel-exporter/internal/exporter/httpexporter"
	"github.com/firebolt-db/otel-exporter/internal/logging"
)

func Test_Config(t *testing.T) {
	os.Clearenv()

	// Set minimal config
	require.NoError(t, errors.Join(
		os.Setenv("FIREBOLT_OTEL_EXPORTER_ACCOUNTS", "acc1,acc2"),
		os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_ID", "client_id"),
		os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET", "client_secret"),
		os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS", "grpc_address"),
	))

	cfg, err := config.NewConfig(context.Background())
	require.NoError(t, err)

	require.Equal(t, &config.Config{
		Logging: logging.Config{
			Format: logging.FormatJSON,
			Level:  logging.LevelInfo,
		},
		Accounts: []string{"acc1", "acc2"},
		Credentials: config.Credentials{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
		},
		Exporter: config.ExporterConfig{
			GRPC: &grpcexporter.Config{
				Address: "grpc_address",
			},
		},
		CollectInterval: 30 * time.Second,
	}, cfg)
}

func Test_Config_MissingCreds(t *testing.T) {
	os.Clearenv()

	os.Setenv("FIREBOLT_OTEL_EXPORTER_ACCOUNTS", "acc1,acc2")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS", "grpc_address")

	cfg, err := config.NewConfig(context.Background())
	require.Error(t, err)
	require.Nil(t, cfg)
}

func Test_Config_OverrideDefaults(t *testing.T) {
	os.Clearenv()

	os.Setenv("FIREBOLT_OTEL_EXPORTER_LOG_FORMAT", "text")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_LOG_LEVEL", "debug")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_COLLECT_INTERVAL", "1m")

	os.Setenv("FIREBOLT_OTEL_EXPORTER_ACCOUNTS", "acc1,acc2")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_ID", "client_id")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET", "client_secret")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS", "grpc_address")

	cfg, err := config.NewConfig(context.Background())
	require.NoError(t, err)

	require.Equal(t, &config.Config{
		Logging: logging.Config{
			Format: logging.FormatText,
			Level:  logging.LevelDebug,
		},
		Accounts: []string{"acc1", "acc2"},
		Credentials: config.Credentials{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
		},
		Exporter: config.ExporterConfig{
			GRPC: &grpcexporter.Config{
				Address: "grpc_address",
			},
		},
		CollectInterval: 60 * time.Second,
	}, cfg)
}

func Test_Config_GRPC(t *testing.T) {
	os.Clearenv()

	os.Setenv("FIREBOLT_OTEL_EXPORTER_ACCOUNTS", "acc1,acc2")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_ID", "client_id")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET", "client_secret")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS", "grpc_address")

	os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_OAUTH_CLIENT_ID", "oauth_client_id")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_OAUTH_CLIENT_SECRET", "oauth_client_secret")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_GRPC_OAUTH_TOKEN_URL", "oauth_token_url")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_SYSTEM_CERT_POOL", "true")

	cfg, err := config.NewConfig(context.Background())
	require.NoError(t, err)

	require.Equal(t, &config.Config{
		Logging: logging.Config{
			Format: logging.FormatJSON,
			Level:  logging.LevelInfo,
		},
		Accounts: []string{"acc1", "acc2"},
		Credentials: config.Credentials{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
		},
		Exporter: config.ExporterConfig{
			GRPC: &grpcexporter.Config{
				Address: "grpc_address",
				Credentials: grpcexporter.ConfigCredentials{
					RPC: grpcexporter.ConfigCredentialsRPC{
						OAuth2: &grpcexporter.ConfigCredentialsRPCOAuth2{
							ClientID:     "oauth_client_id",
							ClientSecret: "oauth_client_secret",
							TokenURL:     "oauth_token_url",
						},
					},
					Transport: grpcexporter.ConfigCredentialsTransport{
						SystemCertPool: &grpcexporter.ConfigCredentialsTransportSystemCertPool{
							Enabled: true,
						},
					},
				},
			},
		},
		CollectInterval: 30 * time.Second,
	}, cfg)
}

func Test_Config_HTTP(t *testing.T) {
	os.Clearenv()

	// Set minimal config
	os.Setenv("FIREBOLT_OTEL_EXPORTER_ACCOUNTS", "acc1,acc2")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_ID", "client_id")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET", "client_secret")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_HTTP_ADDRESS", "http_address")

	os.Setenv("FIREBOLT_OTEL_EXPORTER_HTTP_TLS_X509_CERT_PEM_BLOCK", "cert_pem_block")
	os.Setenv("FIREBOLT_OTEL_EXPORTER_HTTP_TLS_X509_KEY_PEM_BLOCK", "key_pem_block")

	cfg, err := config.NewConfig(context.Background())
	require.NoError(t, err)

	require.Equal(t, &config.Config{
		Logging: logging.Config{
			Format: logging.FormatJSON,
			Level:  logging.LevelInfo,
		},
		Accounts: []string{"acc1", "acc2"},
		Credentials: config.Credentials{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
		},
		Exporter: config.ExporterConfig{
			HTTP: &httpexporter.Config{
				Address: "http_address",
				TLS: &httpexporter.ConfigConnectionOptionsTLS{
					X509KeyPair: httpexporter.X509KeyPair{
						CertPEMBlock: "cert_pem_block",
						KeyPEMBlock:  "key_pem_block",
					},
				},
			},
		},
		CollectInterval: 30 * time.Second,
	}, cfg)
}
