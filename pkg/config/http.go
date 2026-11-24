package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги для настройки HTTP/REST веб-сервера (размеры тела запросов, TLS и таймауты).
//
//nolint:gosec // G101 : это флаги (параметры)
const (
	FlagHTTPAddress            string = "http-address"               // CLI-флаг сетевого адреса HTTP слушателя
	FlagHTTPReadTimeout        string = "http-read-timeout"          // CLI-флаг лимита времени чтения заголовков и боди запроса
	FlagHTTPWriteTimeout       string = "http-write-timeout"         // CLI-флаг лимита времени записи байт ответа клиенту
	FlagHTTPIdleTimeout        string = "http-idle-timeout"          // CLI-флаг лимита времени простоя Keep-Alive HTTP сессий
	FlagHTTPShutdownTimeout    string = "http-shutdown-timeout"      // CLI-флаг таймаута мягкого гашения HTTP сервера
	FlagHTTPPrivateKeyPath     string = "http-private-key-path"      // CLI-флаг пути к приватному ключу шифрования TLS
	FlagHTTPCertificatePath    string = "http-certificate-path"      // CLI-флаг пути к SSL/TLS сертификату веб-сервера
	FlagHTTPSecure             string = "http-secure"                // CLI-флаг активации защищенного режима HTTPS (true/false)
	FlagHTTPMaxRequestBodySize string = "http-max-request-body-size" // CLI-флаг ограничения размера боди (Защита от DOS)
)

// Строковые ключи Viper-маппинга для иерархического связывания параметров HTTP/REST сервера.
//
//nolint:gosec // G101 : это ключи
const (
	KeyHTTPAddress            string = "http.address"
	KeyHTTPReadTimeout        string = "http.read_timeout"
	KeyHTTPWriteTimeout       string = "http.write_timeout"
	KeyHTTPIdleTimeout        string = "http.idle_timeout"
	KeyHTTPShutdownTimeout    string = "http.shutdown_timeout"
	KeyHTTPPrivateKeyPath     string = "http.private_key_path"
	KeyHTTPCertificatePath    string = "http.certificate_path"
	KeyHTTPSecure             string = "http.secure"
	KeyHTTPMaxRequestBodySize string = "http.max_request_body_size"
)

// Константные лимиты и дефолты HTTP/REST веб-сервера.
const (
	DefaultHTTPAddress            string        = "localhost:8080" // Сетевой интерфейс и порт по умолчанию для REST-API
	DefaultHTTPSecure             bool          = false            // Отключение HTTPS по умолчанию (Plain HTTP текстовый режим)
	DefaultHTTPReadTimeout        time.Duration = 5 * time.Second  // Порог ожидания при чтении заголовков и боди входящего запроса
	DefaultHTTPWriteTimeout       time.Duration = 5 * time.Second  // Порог ожидания при передаче байт ответа обратно клиенту
	DefaultHTTPIdleTimeout        time.Duration = 30 * time.Second // Порог удержания сокета в режиме ожидания нового Keep-Alive запроса
	DefaultHTTPShutdownTimeout    time.Duration = 15 * time.Second // Лимит времени на доваривание запросов при Graceful Shutdown сервера
	DefaultHTTPMaxRequestBodySize int           = 1024 * 1024 * 4  // Верхний жесткий лимит тела HTTP запроса — 4MB (Защита от OOM)
)

// HTTPConfig — настройки сервера и таймауты
//
// Инкапсулирует конфигурационные параметры HTTP/REST веб-сервера, включая адресацию,
// лимиты размеров входящего трафика, сетевые таймауты сокетов и параметры шифрования TLS.
type HTTPConfig struct {
	Address            string        `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout" json:"read_timeout,omitempty" yaml:"read_timeout,omitempty" toml:"read_timeout,omitempty"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout" json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty" toml:"idle_timeout,omitempty"`
	ShutdownTimeout    time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty" toml:"shutdown_timeout,omitempty"`
	PrivateKeyPath     string        `mapstructure:"private_key_path" json:"private_key_path,omitempty" yaml:"private_key_path,omitempty" toml:"private_key_path,omitempty"`
	CertificatePath    string        `mapstructure:"certificate_path" json:"certificate_path,omitempty" yaml:"certificate_path,omitempty" toml:"certificate_path,omitempty"`
	Secure             bool          `mapstructure:"secure" json:"secure,omitempty" yaml:"secure,omitempty" toml:"secure,omitempty"`
	MaxRequestBodySize int           `mapstructure:"max_request_body_size" json:"max_request_body_size,omitempty" yaml:"max_request_body_size,omitempty" toml:"max_request_body_size,omitempty"`
}

// NewHTTPConfig конструктор
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

// NewDefaultHTTPConfig конструктор настроек по умолчанию
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

// Validate осуществляет семантическую и системную валидацию параметров HTTP-транспорта на этапе запуска (Bootstrap Phase).
// Полностью пресекает попытки запуска веб-интерфейса с незаданными адресами, несмонтированными TLS-ключами или невалидными таймаутами.
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
		// Проактивно проверяем доступность и целостность ключевой пары в файловой системе (Fail-Fast паттерн)
		if _, err := os.Stat(hc.PrivateKeyPath); err != nil {
			return errs.NewConfigValidateError("http", "private_key_path", "must be a valid path", err)
		}
		if _, err := os.Stat(hc.CertificatePath); err != nil {
			return errs.NewConfigValidateError("http", "certificate_path", "must be a valid path", err)
		}
	}
	if hc.MaxRequestBodySize < 0 {
		return errs.NewConfigValidateError("http", "max_request_body_size", "must be greater or equal than zero", nil)
	}

	return nil
}

// BuildDefaultHTTPSettings http settings builder
//
// # ВНИМАНИЕ!!!
//
// Сюда не входят:
//
// KeyHTTPPrivateKeyPath FlagHTTPPrivateKeyPath
// KeyHTTPCertificatePath FlagHTTPCertificatePath
func BuildDefaultHTTPSettings() Settings {
	return Settings{
		NewSetting(KeyHTTPAddress, FlagHTTPAddress, DefaultHTTPAddress, "http address"),
		NewSetting(KeyHTTPReadTimeout, FlagHTTPReadTimeout, DefaultHTTPReadTimeout, "http read timeout"),
		NewSetting(KeyHTTPWriteTimeout, FlagHTTPWriteTimeout, DefaultHTTPWriteTimeout, "http write timeout"),
		NewSetting(KeyHTTPIdleTimeout, FlagHTTPIdleTimeout, DefaultHTTPIdleTimeout, "http idle timeout"),
		NewSetting(KeyHTTPShutdownTimeout, FlagHTTPShutdownTimeout, DefaultHTTPShutdownTimeout, "http shutdown timeout"),
		//NewSetting(KeyHTTPPrivateKeyPath, FlagHTTPPrivateKeyPath, "", "http private key path"),
		//NewSetting(KeyHTTPCertificatePath, FlagHTTPCertificatePath, "", "http certificate path"),
		NewSetting(KeyHTTPSecure, FlagHTTPSecure, DefaultHTTPSecure, "http secure mode"),
		NewSetting(KeyHTTPMaxRequestBodySize, FlagHTTPMaxRequestBodySize, DefaultHTTPMaxRequestBodySize, "http max request body size"),
	}
}
