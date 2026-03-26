package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// HTTPConfig — настройки сервера и таймауты
type HTTPConfig struct {
	Address            string        `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout" json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout" json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty"`
	ShutdownTimeout    time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"`
	PrivateKeyPath     string        `mapstructure:"private_key_path" json:"private_key_path,omitempty" yaml:"private_key_path,omitempty"`
	CertificatePath    string        `mapstructure:"certificate_path" json:"certificate_path,omitempty" yaml:"certificate_path,omitempty"`
	Secure             bool          `mapstructure:"secure" json:"secure,omitempty" yaml:"secure,omitempty"`
	MaxRequestBodySize int           `mapstructure:"max_request_body_size" json:"max_request_body_size,omitempty" yaml:"max_request_body_size,omitempty"`
}

func NewHTTPConfig(
	address string,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	shutdownTimeout time.Duration,
	privateKeyPath string,
	certificatePath string,
	secure bool,
	maxRequestBodySize int,
) *HTTPConfig {
	return &HTTPConfig{
		Address:            address,
		ReadTimeout:        readTimeout,
		WriteTimeout:       writeTimeout,
		IdleTimeout:        idleTimeout,
		ShutdownTimeout:    shutdownTimeout,
		PrivateKeyPath:     privateKeyPath,
		CertificatePath:    certificatePath,
		Secure:             secure,
		MaxRequestBodySize: maxRequestBodySize,
	}
}

func NewDefaultHTTPConfig() *HTTPConfig {
	return NewHTTPConfig(
		"",
		DefaultHTTPReadTimeout,
		DefaultHTTPWriteTimeout,
		DefaultHTTPIdleTimeout,
		DefaultHTTPShutdownTimeout,
		"",
		"",
		DefaultHTTPSecure,
		DefaultHTTPMaxRequestBodySize,
	)
}

func (hc *HTTPConfig) Validate() error {
	if hc.Address == "" {
		return errs.NewConfigValidateError("http", "address", "must not be empty", nil)
	}
	if hc.ReadTimeout <= 0 {
		return errs.NewConfigValidateError("http", "read_timeout", "must be greater than zero", nil)
	}
	if hc.WriteTimeout <= 0 {
		return errs.NewConfigValidateError("http", "write_timeout", "must be greater than zero", nil)
	}
	if hc.IdleTimeout <= 0 {
		return errs.NewConfigValidateError("http", "idle_timeout", "must be greater than zero", nil)
	}
	if hc.ShutdownTimeout <= 0 {
		return errs.NewConfigValidateError("http", "shutdown_timeout", "must be greater than zero", nil)
	}
	if hc.Secure {
		if hc.PrivateKeyPath == "" {
			return errs.NewConfigValidateError("http", "private_key_path", "must not be empty", nil)
		}
		if hc.CertificatePath == "" {
			return errs.NewConfigValidateError("http", "certificate_path", "must not be empty", nil)
		}
		if _, err := os.Stat(hc.PrivateKeyPath); err != nil {
			return errs.NewConfigValidateError("http", "private_key_path", "must be a valid path", err)
		}
		if _, err := os.Stat(hc.CertificatePath); err != nil {
			return errs.NewConfigValidateError("http", "certificate_path", "must be a valid path", err)
		}
	}
	if !(hc.MaxRequestBodySize >= 0) {
		return errs.NewConfigValidateError("http", "max_request_body_size", "must be greater or equal than zero", nil)
	}

	return nil
}
