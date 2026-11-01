package config

import (
	"fmt"
	"os"

	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
)

func initFLags() (res *pflag.FlagSet, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Проверяем, является ли r ошибкой
			recoveryErr, ok := r.(error)
			if !ok {
				// Если это строка или что-то другое, приводим к виду error вручную
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}
			res = nil
			err = errs.NewConfigError("parse cli flags panic", recoveryErr)
		}
	}()

	res = pflag.NewFlagSet("cmd flags", pflag.PanicOnError)

	// Используем константы Flag...
	res.String(conf.FlagConfig, "config/config.yaml", "path to config file")
	res.String(conf.FlagAppEnv, string(conf.DefaultAppEnv), "application environment (prod, dev, test)")
	res.Duration(conf.FlagAppInitTimeout, conf.DefaultAppInitTimeout, "application init timeout")
	res.Duration(conf.FlagAppStopTimeout, conf.DefaultAppStopTimeout, "application stop timeout")
	res.Duration(conf.FlagAppCloseTimeout, conf.DefaultAppCloseTimeout, "application close timeout")
	res.String(flagAppNodeName, defaultNodeName, "application node name")

	// Auth
	res.String(conf.FlagAuthJWTSecret, "", "JWT secret")
	res.String(conf.FlagAuthJWTSigningMethod, conf.DefaultAuthSigningMethod, "JWT signing method")
	res.Duration(conf.FlagAuthAccessTokenTTL, conf.DefaultAuthAccessTokenTTL, "JWT access token TTL")
	res.Duration(conf.FlagAuthRefreshTokenTTL, conf.DefaultAuthRefreshTokenTTL, "JWT refresh token TTL")
	res.String(conf.FlagAuthRSAPrivateKeyPath, "", "RSA private key path")
	res.String(conf.FlagAuthMasterPasswordSalt, "", "master password salt")

	// HTTP
	res.String(conf.FlagHTTPAddress, conf.DefaultHTTPAddress, "http address")
	res.Duration(conf.FlagHTTPReadTimeout, conf.DefaultHTTPReadTimeout, "http read timeout")
	res.Duration(conf.FlagHTTPWriteTimeout, conf.DefaultHTTPWriteTimeout, "http write timeout")
	res.Duration(conf.FlagHTTPIdleTimeout, conf.DefaultHTTPIdleTimeout, "http idle timeout")
	res.Duration(conf.FlagHTTPShutdownTimeout, conf.DefaultHTTPShutdownTimeout, "http shutdown timeout")
	res.String(conf.FlagHTTPPrivateKeyPath, "", "http private key path")
	res.String(conf.FlagHTTPCertificatePath, "", "http certificate path")
	res.Bool(conf.FlagHTTPSecure, conf.DefaultHTTPSecure, "http secure mode")
	res.Int(conf.FlagHTTPMaxRequestBodySize, conf.DefaultHTTPMaxRequestBodySize, "http max request body size")

	// gRPC
	res.String(conf.FlagGRPCAddress, conf.DefaultGRPCAddress, "gRPC address")
	res.Duration(conf.FlagGRPCMaxConnIdle, conf.DefaultGRPCMaxConnIdle, "gRPC max connection idle timeout")
	res.Duration(conf.FlagGRPCMaxConnAge, conf.DefaultGRPCMaxConnAge, "gRPC max connection age timeout")
	res.Duration(conf.FlagGRPCMaxConnAgeGrace, conf.DefaultGRPCMaxConnAgeGrace, "gRPC max connection age grace timeout")
	res.Duration(conf.FlagGRPCTimeout, conf.DefaultGRPCTimeout, "gRPC timeout")
	res.Duration(conf.FlagGRPCKeepAliveTime, conf.DefaultGRPCKeepAliveTime, "gRPC keep alive timeout")
	res.Duration(conf.FlagGRPCKeepAliveTimeout, conf.DefaultGRPCKeepAliveTimeout, "gRPC keep alive timeout")
	res.Duration(conf.FlagGRPCShutdownTimeout, conf.DefaultGRPCShutdownTimeout, "gRPC shutdown timeout")

	// DB
	res.String(conf.FlagDBDSN, conf.DefaultDBDSN, "database dsn")
	res.String(conf.FlagDBDriver, conf.DefaultDBDriver, "database driver name/alias")
	res.Int(conf.FlagDBMaxOpenConns, conf.DefaultDBMaxOpenConns, "db max open connections")
	res.Int(conf.FlagDBMaxIdleConns, conf.DefaultDBMaxIdleConns, "db max idle connections")
	res.Duration(conf.FlagDBMaxIdleLifetime, conf.DefaultDBConnMaxIdleLifetime, "db max idle connection lifetime")
	res.Duration(conf.FlagDBConnTimeout, conf.DefaultDBConnTimeout, "db connection timeout)")

	// Log
	res.String(conf.FlagLogLevel, conf.DefaultLogLevel, "log level")
	res.String(conf.FlagLogFormat, conf.DefaultLogFormat, "log format")

	// Telemetry
	res.Bool(conf.FlagTelemetryEnabled, conf.DefaultTelemetryEnabled, "telemetry enabled")
	res.String(conf.FlagTelemetryServiceName, "", "telemetry service name")
	res.String(conf.FlagTelemetryExporterEndpoint, conf.DefaultTelemetryExporterEndpoint, "telemetry exporter endpoint")
	res.Float64(conf.FlagTelemetrySampleRate, conf.DefaultTelemetrySampleRate, "telemetry sample rate")
	res.Duration(conf.FlagTelemetryTimeout, conf.DefaultTelemetryTimeout, "telemetry timeout")

	// Добавь остальные pflag.String/Int/Duration для Redis, etc. ...
	// ..

	// Парсинг
	err = res.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return res, err
}
