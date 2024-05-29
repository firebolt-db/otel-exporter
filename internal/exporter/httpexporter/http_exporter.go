package httpexporter

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
)

// NewHTTPExporter creates a new instance of otlpmetrichttp.Exporter
func NewHTTPExporter(ctx context.Context, cfg *Config) (*otlpmetrichttp.Exporter, error) {
	// configure TLS
	var tlsConfig *tls.Config
	if cfg.TLS != nil {
		cert, err := tls.X509KeyPair(
			[]byte(cfg.TLS.X509KeyPair.CertPEMBlock),
			[]byte(cfg.TLS.X509KeyPair.KeyPEMBlock),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read X509KeyPair: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	var opts = []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.Address),
	}

	if tlsConfig != nil {
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	} else {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)

	if err != nil {
		return nil, err
	}

	return exporter, nil
}
