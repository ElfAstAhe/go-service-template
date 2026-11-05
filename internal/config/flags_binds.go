package config

import (
	"errors"

	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindFlags(flags *pflag.FlagSet, v *viper.Viper) error {
	err := errors.Join(
		// App
		v.BindPFlag(conf.KeyAppEnv, flags.Lookup(conf.FlagAppEnv)),
		v.BindPFlag(conf.KeyAppInitTimeout, flags.Lookup(conf.FlagAppInitTimeout)),
		v.BindPFlag(conf.KeyAppStopTimeout, flags.Lookup(conf.FlagAppStopTimeout)),
		v.BindPFlag(conf.KeyAppCloseTimeout, flags.Lookup(conf.FlagAppCloseTimeout)),
		v.BindPFlag(keyAppNodeName, flags.Lookup(flagAppNodeName)),
		// Auth
		v.BindPFlag(conf.KeyAuthJWTSecret, flags.Lookup(conf.FlagAuthJWTSecret)),
		v.BindPFlag(conf.KeyAuthJWTSigningMethod, flags.Lookup(conf.FlagAuthJWTSigningMethod)),
		v.BindPFlag(conf.KeyAuthAccessTokenTTL, flags.Lookup(conf.FlagAuthAccessTokenTTL)),
		v.BindPFlag(conf.KeyAuthRefreshTokenTTL, flags.Lookup(conf.FlagAuthRefreshTokenTTL)),
		v.BindPFlag(conf.KeyAuthRSAPrivateKeyPath, flags.Lookup(conf.FlagAuthRSAPrivateKeyPath)),
		v.BindPFlag(conf.KeyAuthMasterPasswordSalt, flags.Lookup(conf.FlagAuthMasterPasswordSalt)),
		// HTTP
		v.BindPFlag(conf.KeyHTTPAddress, flags.Lookup(conf.FlagHTTPAddress)),
		v.BindPFlag(conf.KeyHTTPReadTimeout, flags.Lookup(conf.FlagHTTPReadTimeout)),
		v.BindPFlag(conf.KeyHTTPWriteTimeout, flags.Lookup(conf.FlagHTTPWriteTimeout)),
		v.BindPFlag(conf.KeyHTTPIdleTimeout, flags.Lookup(conf.FlagHTTPIdleTimeout)),
		v.BindPFlag(conf.KeyHTTPShutdownTimeout, flags.Lookup(conf.FlagHTTPShutdownTimeout)),
		v.BindPFlag(conf.KeyHTTPPrivateKeyPath, flags.Lookup(conf.FlagHTTPPrivateKeyPath)),
		v.BindPFlag(conf.KeyHTTPCertificatePath, flags.Lookup(conf.FlagHTTPCertificatePath)),
		v.BindPFlag(conf.KeyHTTPSecure, flags.Lookup(conf.FlagHTTPSecure)),
		v.BindPFlag(conf.KeyHTTPMaxRequestBodySize, flags.Lookup(conf.FlagHTTPMaxRequestBodySize)),
		// gRPC
		v.BindPFlag(conf.KeyGRPCAddress, flags.Lookup(conf.FlagGRPCAddress)),
		v.BindPFlag(conf.KeyGRPCMaxConnIdle, flags.Lookup(conf.FlagGRPCMaxConnIdle)),
		v.BindPFlag(conf.KeyGRPCMaxConnAge, flags.Lookup(conf.FlagGRPCMaxConnAge)),
		v.BindPFlag(conf.KeyGRPCMaxConnAgeGrace, flags.Lookup(conf.FlagGRPCMaxConnAgeGrace)),
		v.BindPFlag(conf.KeyGRPCTimeout, flags.Lookup(conf.FlagGRPCTimeout)),
		v.BindPFlag(conf.KeyGRPCKeepAliveTime, flags.Lookup(conf.FlagGRPCKeepAliveTime)),
		v.BindPFlag(conf.KeyGRPCKeepAliveTimeout, flags.Lookup(conf.FlagGRPCKeepAliveTimeout)),
		v.BindPFlag(conf.KeyGRPCShutdownTimeout, flags.Lookup(conf.FlagGRPCShutdownTimeout)),
		// Log
		v.BindPFlag(conf.KeyLogLevel, flags.Lookup(conf.FlagLogLevel)),
		v.BindPFlag(conf.KeyLogFormat, flags.Lookup(conf.FlagLogFormat)),
		// DB
		v.BindPFlag(conf.KeyDBDriver, flags.Lookup(conf.FlagDBDriver)),
		v.BindPFlag(conf.KeyDBDSN, flags.Lookup(conf.FlagDBDSN)),
		v.BindPFlag(conf.KeyDBMaxOpenConns, flags.Lookup(conf.FlagDBMaxOpenConns)),
		v.BindPFlag(conf.KeyDBMaxIdleConns, flags.Lookup(conf.FlagDBMaxIdleConns)),
		v.BindPFlag(conf.KeyDBConnMaxIdleLifetime, flags.Lookup(conf.FlagDBMaxIdleLifetime)),
		v.BindPFlag(conf.KeyDBConnTimeout, flags.Lookup(conf.FlagDBConnTimeout)),
		// Telemetry
		v.BindPFlag(conf.KeyTelemetryEnabled, flags.Lookup(conf.FlagTelemetryEnabled)),
		v.BindPFlag(conf.KeyTelemetryExporterEndpoint, flags.Lookup(conf.FlagTelemetryExporterEndpoint)),
		v.BindPFlag(conf.KeyTelemetryServiceName, flags.Lookup(conf.FlagTelemetryServiceName)),
		v.BindPFlag(conf.KeyTelemetrySampleRate, flags.Lookup(conf.FlagTelemetrySampleRate)),
		v.BindPFlag(conf.KeyTelemetryTimeout, flags.Lookup(conf.FlagTelemetryTimeout)),
	)
	if err != nil {
		return errs.NewConfigError("bind flags with keys", err)
	}

	return nil
}
