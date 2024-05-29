package grpcexporter

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
)

type Config struct {
	Address     string `env:"FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS"`
	Credentials ConfigCredentials
}

func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Address, validation.Required),
		validation.Field(&c.Credentials),
	)
}

func (c Config) DialOptions() ([]grpc.DialOption, error) {
	credsOpts, err := c.Credentials.DialOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build credentials dial options: %w", err)
	}

	return credsOpts, nil
}

// ConfigCredentials represents gRPC authentication settings.
type ConfigCredentials struct {

	// RPC sets credentials and places auth state on each outbound RPC
	RPC ConfigCredentialsRPC

	// Transport configures a connection level security credentials (e.g., TLS/SSL)
	Transport ConfigCredentialsTransport
}

// DialOptions returns a slice of grpc.DialOption based on configuration values.
func (c ConfigCredentials) DialOptions() ([]grpc.DialOption, error) {

	rpcOpts, err := c.RPC.DialOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build RPC dial options: %w", err)
	}

	transportOpts, err := c.Transport.DialOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build trasport dial options: %w", err)
	}

	return append(rpcOpts, transportOpts...), nil
}

// ConfigCredentialsRPC represents client authentication settings.
type ConfigCredentialsRPC struct {

	// OAuth2 configures OAuth2 client authentication.
	OAuth2 *ConfigCredentialsRPCOAuth2
}

// DialOptions returns a slice of grpc.DialOption based on configuration values.
func (c ConfigCredentialsRPC) DialOptions() ([]grpc.DialOption, error) {

	if c.OAuth2 != nil {
		return c.OAuth2.DialOptions()
	}

	return nil, nil
}

// ConfigCredentialsRPCOAuth2 represents OAuth2 client authentication settings.
type ConfigCredentialsRPCOAuth2 struct {

	// ClientID is the application's ID.
	ClientID string `env:"FIREBOLT_OTEL_EXPORTER_GRPC_OAUTH_CLIENT_ID"`

	// ClientSecret is the application's secret.
	ClientSecret string `env:"FIREBOLT_OTEL_EXPORTER_GRPC_OAUTH_CLIENT_SECRET"`

	// TokenURL is the resource server's token endpoint URL.
	// This is a constant specific to each server.
	TokenURL string `env:"FIREBOLT_OTEL_EXPORTER_GRPC_OAUTH_TOKEN_URL"`
}

// DialOptions returns a slice of grpc.DialOption based on configuration values.
func (c ConfigCredentialsRPCOAuth2) DialOptions() ([]grpc.DialOption, error) {
	c2 := clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TokenURL:     c.TokenURL,
	}
	return []grpc.DialOption{
		grpc.WithPerRPCCredentials(
			oauth.TokenSource{
				TokenSource: c2.TokenSource(
					context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
						Timeout: 5 * time.Second,
					}),
				),
			},
		),
	}, nil
}

// ConfigCredentialsTransport represents transport authentication settings.
type ConfigCredentialsTransport struct {
	// SystemCertPool enables TLS security based on operating system certificate pool.
	SystemCertPool *ConfigCredentialsTransportSystemCertPool `env:",noinit"`
}

// DialOptions returns a slice of grpc.DialOption based on configuration values.
func (c ConfigCredentialsTransport) DialOptions() ([]grpc.DialOption, error) {
	if c.SystemCertPool != nil {
		return c.SystemCertPool.DialOptions()
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, nil
}

// ConfigCredentialsTransportSystemCertPool enables TLS security based on operating system certificate pool.
type ConfigCredentialsTransportSystemCertPool struct {
	Enabled bool `env:"FIREBOLT_OTEL_EXPORTER_SYSTEM_CERT_POOL,default=false"`
}

// DialOptions returns a slice of grpc.DialOption based on configuration values.
func (c ConfigCredentialsTransportSystemCertPool) DialOptions() ([]grpc.DialOption, error) {
	if !c.Enabled {
		return nil, nil
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to load system cert pool: %w", err)
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{RootCAs: certPool}),
		),
	}, nil
}
