package config

import (
	"time"
)

const (
	defaultAppEnv = AppEnvDevelopment
)

const (
	keyAppEnv string = "app.env"
)

// HTTP defaults
const (
	defaultHTTPAddress         string        = "localhost:8080"
	defaultHTTPSecure          bool          = false
	defaultHTTPReadTimeout     time.Duration = 5 * time.Second
	defaultHTTPWriteTimeout    time.Duration = 5 * time.Second
	defaultHTTPIdleTimeout     time.Duration = 30 * time.Second
	defaultHTTPShutdownTimeout time.Duration = 15 * time.Second
)

const (
	keyHTTPAddress         string = "http.address"
	keyHTTPSecure          string = "http.secure"
	keyHTTPReadTimeout     string = "http.read_timeout"
	keyHTTPWriteTimeout    string = "http.write_timeout"
	keyHTTPIdleTimeout     string = "http.idle_timeout"
	keyHTTPShutdownTimeout string = "http.shutdown_timeout"
)

// gRPC defaults
const (
	defaultGRPCAddress string        = "localhost:50051"
	defaultGRPCTimeout time.Duration = 5 * time.Second
)

const (
	keyGRPCAddress string = "grpc.address"
	keyGRPCTimeout string = "grpc.timeout"
)

// logger defaults
const (
	defaultLogLevel  string = "info"
	defaultLogFormat string = "console"
)

const (
	keyLogLevel  string = "log.level"
	keyLogFormat string = "log.format"
)

// DB defaults (only pool settings)
const (
	defaultDBMaxOpenConns        int           = 32
	defaultDBMaxIdleConns        int           = 4
	defaultDBConnMaxIdleLifetime time.Duration = 60 * time.Second
	defaultDBConnTimeout         time.Duration = 30 * time.Second
)

const (
	keyDBMaxOpenConns        string = "db.max_open_conns"
	keyDBMaxIdleConns        string = "db.max_idle_conns"
	keyDBConnMaxIdleLifetime string = "db.conn_max_idle_lifetime"
	keyDBConnTimeout         string = "db.conn_timeout"
)
