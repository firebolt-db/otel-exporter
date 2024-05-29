package grpcexporter

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"google.golang.org/grpc"
)

func NewGRPCExporter(ctx context.Context, cfg *Config) (*otlpmetricgrpc.Exporter, error) {
	dialOpts, err := cfg.DialOptions()
	if err != nil {
		return nil, err
	}

	dialOpts = append(dialOpts, grpc.WithContextDialer(
		func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
		}),
	)

	conn, err := grpc.NewClient(cfg.Address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
	)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
