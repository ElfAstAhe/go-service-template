package config

type AppEnv string

func (ae AppEnv) Exists() bool {
	return appEnvs.Contains(ae)
}

type appEnvList map[AppEnv]struct{}

func (ae appEnvList) Contains(env AppEnv) bool {
	_, ok := ae[env]

	return ok
}

// app env enum
const (
	AppEnvProduction  AppEnv = "prod"
	AppEnvDevelopment AppEnv = "dev"
	AppEnvTest        AppEnv = "test"
)

var appEnvs appEnvList = map[AppEnv]struct{}{
	AppEnvProduction:  {},
	AppEnvDevelopment: {},
	AppEnvTest:        {},
}
