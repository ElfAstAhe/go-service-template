package config

// AppEnv определяет строго типизированный строковый доменный тип для сред окружения (Environment Type).
type AppEnv string

// Exists выполняет атомарную проверку за константное время O(1), входит ли текущая строка
// в пул легитимных сред окружения платформы, защищая рантайм от некорректных конфигураций.
func (ae AppEnv) Exists() bool {
	return appEnvs.Contains(ae)
}

// appEnvList инкапсулирует внутреннюю структуру реестра доступных сред окружения.
type appEnvList map[AppEnv]struct{}

// Contains верифицирует фактическое наличие переданного ключа в хэш-карте.
func (ae appEnvList) Contains(env AppEnv) bool {
	_, ok := ae[env]

	return ok
}

// Набор констант поддерживаемых и валидируемых сред окружения (App Env Enum).
const (
	// AppEnvProduction промышленный контур (Production)
	AppEnvProduction AppEnv = "prod"

	// AppEnvDevelopment контур локальной разработки и отладки (Development)
	AppEnvDevelopment AppEnv = "dev"

	// AppEnvTest контур автоматизированного тестирования (CI/CD / Testing)
	AppEnvTest AppEnv = "test"
)

// Глобальный неизменяемый синглтон-реестр валидных окружений.
// 💡 Оптимизация: пустая структура struct{} имеет нулевой размер (0 bytes), исключая нагрузку на GC.
var appEnvs appEnvList = map[AppEnv]struct{}{
	AppEnvProduction:  {},
	AppEnvDevelopment: {},
	AppEnvTest:        {},
}
