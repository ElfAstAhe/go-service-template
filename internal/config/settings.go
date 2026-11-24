package config

import (
	pkgconf "github.com/ElfAstAhe/go-service-template/pkg/config"
)

func buildCommonSettings() pkgconf.Settings {
	return pkgconf.Settings{
		pkgconf.NewSetting("", pkgconf.FlagConfig, pkgconf.DefaultConfigFile, "application config file path"),
	}
}

func buildHTTPSettings() pkgconf.Settings {
	return append(pkgconf.BuildDefaultHTTPSettings(),
		pkgconf.NewSetting(pkgconf.KeyHTTPPrivateKeyPath, pkgconf.FlagHTTPPrivateKeyPath, "", "http private key path"),
		pkgconf.NewSetting(pkgconf.KeyHTTPCertificatePath, pkgconf.FlagHTTPCertificatePath, "", "http certificate path"),
	)
}

func buildDBSettings() pkgconf.Settings {
	return append(pkgconf.BuildDefaultDBSettings(),
		pkgconf.NewSetting(pkgconf.KeyDBDSN, pkgconf.FlagDBDSN, "", "DB connection string"),
	)
}

func buildTelemetrySettings() pkgconf.Settings {
	return append(pkgconf.BuildDefaultTelemetrySettings(),
		pkgconf.NewSetting(pkgconf.KeyTelemetryServiceName, pkgconf.FlagTelemetryServiceName, defaultNodeName, "telemetry service name"),
	)
}
