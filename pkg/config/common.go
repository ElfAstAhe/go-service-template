package config

import (
	"fmt"
)

// Validator поддержка метода проверки конфигурации
type Validator interface {
	Validate() error
}

// Набор констант поддерживаемых и валидируемых сред окружения (App Env Enum).
const (
	// AppEnvProduction промышленный контур (Production)
	AppEnvProduction string = "prod"

	// AppEnvDevelopment контур локальной разработки и отладки (Development)
	AppEnvDevelopment string = "dev"

	// AppEnvTest контур автоматизированного тестирования (CI/CD / Testing)
	AppEnvTest string = "test"
)

// Глобальный неизменяемый синглтон-реестр валидных окружений.
// 💡 Оптимизация: пустая структура struct{} имеет нулевой размер (0 bytes), исключая нагрузку на GC.
var appEnvs = map[string]struct{}{
	AppEnvProduction:  {},
	AppEnvDevelopment: {},
	AppEnvTest:        {},
}

// ValidateAppEnv validate environment parameter value
func ValidateAppEnv(env string) error {
	if _, ok := appEnvs[env]; !ok {
		return fmt.Errorf("invalid environment %q", env)
	}

	return nil
}
