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
	KeyHTTPSecure             string = "http.secure"
	KeyHTTPReadTimeout        string = "http.read_timeout"
	KeyHTTPWriteTimeout       string = "http.write_timeout"
	KeyHTTPIdleTimeout        string = "http.idle_timeout"
	KeyHTTPShutdownTimeout    string = "http.shutdown_timeout"
	KeyHTTPMaxRequestBodySize string = "http.max_request_body_size"
)

// gRPC defaults
const (
	DefaultGRPCAddress         string        = "localhost:50051"
	DefaultGRPCTimeout         time.Duration = 5 * time.Second
	DefaultGRPCShutdownTimeout time.Duration = 15 * time.Second
)

const (
	KeyGRPCAddress         string = "grpc.address"
	KeyGRPCTimeout         string = "grpc.timeout"
	KeyGRPCShutdownTimeout string = "grpc.shutdown_timeout"
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
	DefaultDBMaxOpenConns        int           = 32
	DefaultDBMaxIdleConns        int           = 4
	DefaultDBConnMaxIdleLifetime time.Duration = 60 * time.Second
	DefaultDBConnTimeout         time.Duration = 30 * time.Second
)

const (
	KeyDBMaxOpenConns        string = "db.max_open_conns"
	KeyDBMaxIdleConns        string = "db.max_idle_conns"
	KeyDBConnMaxIdleLifetime string = "db.conn_max_idle_lifetime"
	KeyDBConnTimeout         string = "db.conn_timeout"
)
