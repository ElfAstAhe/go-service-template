package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// HTTPConfig — настройки сервера и таймауты
type HTTPConfig struct {
	Address            string        `mapstructure:"address"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout    time.Duration `mapstructure:"shutdown_timeout"`
	PrivateKeyPath     string        `mapstructure:"private_key_path"`
	CertificatePath    string        `mapstructure:"certificate_path"`
	Secure             bool          `mapstructure:"secure"`
	MaxRequestBodySize int           `mapstructure:"max_request_body_size"`
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
