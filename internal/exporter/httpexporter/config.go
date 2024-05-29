package httpexporter

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Address string `env:"FIREBOLT_OTEL_EXPORTER_HTTP_ADDRESS"`

	TLS *ConfigConnectionOptionsTLS `env:",noinit"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Address, validation.Required),
		validation.Field(&c.TLS),
	)
}

type ConfigConnectionOptionsTLS struct {

	// X509KeyPair to use for mTLS authentication.
	X509KeyPair X509KeyPair
}

// X509KeyPair represents X509 key pair used for mTLS authentication.
type X509KeyPair struct {

	// CertPEMBlock specifies TLS certificate PEM.
	CertPEMBlock string `env:"FIREBOLT_OTEL_EXPORTER_HTTP_TLS_X509_CERT_PEM_BLOCK"`

	// KeyPEMBlock specifies TLS key PEM.
	KeyPEMBlock string `env:"FIREBOLT_OTEL_EXPORTER_HTTP_TLS_X509_KEY_PEM_BLOCK"`
}

// Validate ensures that config is valid.
func (c X509KeyPair) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.CertPEMBlock, validation.Required),
		validation.Field(&c.KeyPEMBlock, validation.Required),
	)
}
