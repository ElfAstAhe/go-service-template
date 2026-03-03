package config

import (
	"time"
)

// HTTP defaults
const (
	DefaultHTTPAddress            string        = "localhost:8080"
	DefaultHTTPSecure             bool          = false
	DefaultHTTPReadTimeout        time.Duration = 5 * time.Second
	DefaultHTTPWriteTimeout       time.Duration = 5 * time.Second
	DefaultHTTPIdleTimeout        time.Duration = 30 * time.Second
	DefaultHTTPShutdownTimeout    time.Duration = 15 * time.Second
	DefaultHTTPMaxRequestBodySize int           = 1024 * 1024 * 4
)

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

// gRPC defaults
const (
	DefaultGRPCAddress string = "localhost:50051"
	// DefaultGRPCMaxConnIdle - Даем соединениям «отдохнуть», но не убиваем их сразу
	DefaultGRPCMaxConnIdle time.Duration = 5 * time.Minute
	// DefaultGRPCMaxConnAge - Ротируем соединения раз в 20 минут для балансировки трафика
	DefaultGRPCMaxConnAge time.Duration = 20 * time.Minute
	// DefaultGRPCMaxConnAgeGrace - Grace-период, чтобы старые запросы успели довариться при закрытии коннекта
	DefaultGRPCMaxConnAgeGrace  time.Duration = 1 * time.Minute
	DefaultGRPCTimeout          time.Duration = 5 * time.Second
	DefaultGRPCKeepAliveTime    time.Duration = 2 * time.Minute
	DefaultGRPCKeepAliveTimeout time.Duration = 20 * time.Second
	DefaultGRPCShutdownTimeout  time.Duration = 15 * time.Second
)

const (
	KeyGRPCAddress          string = "grpc.address"
	KeyGRPCMaxConnIdle      string = "grpc.max-conn-idle"
	KeyGRPCMaxConnAge       string = "grpc.max-conn-age"
	KeyGRPCMaxConnAgeGrace  string = "grpc.max-conn-age-grace"
	KeyGRPCTimeout          string = "grpc.timeout"
	KeyGRPCKeepAliveTime    string = "grpc.keep-alive-time"
	KeyGRPCKeepAliveTimeout string = "grpc.keep-alive-timeout"
	KeyGRPCShutdownTimeout  string = "grpc.shutdown_timeout"
)

// logger defaults
const (
	DefaultLogLevel  string = "info"
	DefaultLogFormat string = "console"
)

const (
	KeyLogLevel  string = "log.level"
	KeyLogFormat string = "log.format"
)

// DB defaults (only pool settings)
const (
	DefaultDBDriver              string        = ""
	DefaultDBDSN                 string        = ""
	DefaultDBMaxOpenConns        int           = 32
	DefaultDBMaxIdleConns        int           = 4
	DefaultDBConnMaxIdleLifetime time.Duration = 60 * time.Second
	DefaultDBConnTimeout         time.Duration = 30 * time.Second
)

const (
	KeyDBDriver              string = "db.driver"
	KeyDBDSN                 string = "db.dsn"
	KeyDBMaxOpenConns        string = "db.max_open_conns"
	KeyDBMaxIdleConns        string = "db.max_idle_conns"
	KeyDBConnMaxIdleLifetime string = "db.conn_max_idle_lifetime"
	KeyDBConnTimeout         string = "db.conn_timeout"
)

// Telemetry defaults
const (
	DefaultTelemetryEnabled          bool          = false
	DefaultTelemetryExporterEndpoint string        = "localhost:4317"
	DefaultTelemetrySampleRate       float64       = 1.0
	DefaultTelemetryTimeout          time.Duration = 5 * time.Second
)

const (
	KeyTelemetryEnabled          string = "telemetry.enabled"
	KeyTelemetryServiceName      string = "telemetry.service_name"
	KeyTelemetryExporterEndpoint string = "telemetry.exporter_endpoint"
	KeyTelemetrySampleRate       string = "telemetry.sample_rate"
	KeyTelemetryTimeout          string = "telemetry.timeout"
)

// Auth
const (
	KeyAuthJWTSecret          string = "auth.jwt_secret"
	KeyAuthAccessTokenTTL     string = "auth.access_token_ttl"
	KeyAuthRefreshTokenTTL    string = "auth.refresh_token_ttl"
	KeyAuthRSAPrivateKeyPath  string = "auth.rsa_private_key_path"
	KeyAuthMasterPasswordSalt string = "auth.master_password_salt"
)
