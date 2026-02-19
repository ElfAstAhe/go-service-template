package config

// 1. Defaults (Значения по умолчанию) — Самый низкий
// Прописаны жестко в коде (твой файл defaults.go). Это фундамент. Если всё остальное отсутствует, сервис
// запустится на этих настройках.
// 2. Config File (YAML/JSON)
// Настройки для конкретного окружения (dev.yaml, prod.yaml). Перекрывают дефолты. Обычно здесь лежат «статичные»
// параметры: таймауты, лимиты пулов.
// 3. Environment Variables (ENV)
// Настройки от инфраструктуры (Docker, Kubernetes). Перекрывают и дефолты, и файл.
// Почему выше файла? Потому что в контейнерах пароли к БД и секреты (JWT) часто прокидываются именно через
// ENV или Secrets. Это позволяет использовать один и тот же YAML-файл для разных сред.
// 4. Flags (Флаги запуска) — Самый высокий
// Параметры, переданные при старте: ./app --port=9090.
// Почему выше всех? Это инструмент «последнего слова» для разработчика или админа. Если нужно быстро
// переопределить порт или путь к конфигу, не меняя ENV и файлы, флаг — единственный способ.

// FlagConfig - файл конфигурации
const FlagConfig = "config-path"

// FlagAppEnv App config flags
const (
	FlagAppEnv string = "env"
)

// Auth config flags
const (
	FlagAuthJWTSecret          string = "auth-jwt-secret"
	FlagAuthAccessTokenTTL     string = "auth-access-token-ttl"
	FlagAuthRefreshTokenTTL    string = "auth-refresh-token-ttl"
	FlagAuthRSAPrivateKeyPath  string = "auth-rsa-private-key-path"
	FlagAuthMasterPasswordSalt string = "auth-master-password-salt"
)

// DB config flags
const (
	FlagDBDSN             string = "db-dsn"
	FlagDBMaxOpenConns    string = "db-max-open-conns"
	FlagDBMaxIdleConns    string = "db-max-idle-conns"
	FlagDBMaxIdleLifetime string = "db-max-idle-lifetime"
	FlagDBConnTimeout     string = "db-conn-timeout"
)

// gRPC config flags
const (
	FlagGRPCAddress          string = "grpc-address"
	FlagGRPCMaxConnIdle      string = "grpc-max-conn-idle"
	FlagGRPCMaxConnAge       string = "grpc-max-conn-age"
	FlagGRPCTimeout          string = "grpc-timeout"
	FlagGRPCKeepAliveTime    string = "grpc-keep-alive-time"
	FlagGRPCKeepAliveTimeout string = "grpc-keep-alive-timeout"
)

// http config flags
const (
	FlagHTTPAddress            string = "http-address"
	FlagHTTPReadTimeout        string = "http-read-timeout"
	FlagHTTPWriteTimeout       string = "http-write-timeout"
	FlagHTTPIdleTimeout        string = "http-idle-timeout"
	FlagHTTPShutdownTimeout    string = "http-shutdown-timeout"
	FlagHTTPPrivateKeyPath     string = "http-private-key-path"
	FlagHTTPCertificatePath    string = "http-certificate-path"
	FlagHTTPSecure             string = "http-secure"
	FlagHTTPMaxRequestBodySize string = "http-max-request-body-size"
)

// log config flags
const (
	FlagLogLevel  string = "log-level"
	FlagLogFormat string = "log-format"
)

// redis config flags
const (
	FlagRedisHost     string = "redis-host"
	FlagRedisPort     string = "redis-port"
	FlagRedisPassword string = "redis-password"
	FlagRedisDB       string = "redis-db"
)

// EnvConfig - файл конфигурации
const EnvConfig string = "CONFIG_PATH"

// Auth config envs
const (
	EnvAuthJWTSecret       string = "AUTH_JWT_SECRET"
	EnvAuthAccessTokenTTL  string = "AUTH_ACCESS_TOKEN_TTL"
	EnvAuthRefreshTokenTTL string = "AUTH_REFRESH_TOKEN_TTL"
	EnvAuthRSAPrivateKey   string = "AUTH_RSAPRIVATE_KEY"
	EnvAuthMasterPassword  string = "AUTH_MASTER_PASSWORD"
)

// DB config envs
const (
	EnvDBDSN             string = "DB_DSN"
	EnvDBMaxOpenConns    string = "DB_MAX_OPEN_CONNS"
	EnvDBMaxIdleConns    string = "DB_MAX_IDLE_CONNS"
	EnvDBMaxIdleLifetime string = "DB_MAX_IDLE_LIFETIME"
	EnvDBConnTimeout     string = "DB_CONN_TIMEOUT"
)

// gRPC config envs
const (
	EnvGRPCAddress          string = "GRPC_ADDRESS"
	EnvGRPCMaxConnIdle      string = "GRPC_MAX_CONN_IDLE"
	EnvGRPCMaxConnAge       string = "GRPC_MAX_CONN_AGE"
	EnvGRPCTimeout          string = "GRPC_TIMEOUT"
	EnvGRPCKeepAliveTime    string = "GRPC_KEEP_ALIVE_TIME"
	EnvGRPCKeepAliveTimeout string = "GRPC_KEEP_ALIVE_TIMEOUT"
)

// HTTP config envs
const (
	EnvHTTPAddress            string = "HTTP_ADDRESS"
	EnvHTTPReadTimeout        string = "HTTP_READ_TIMEOUT"
	EnvHTTPWriteTimeout       string = "HTTP_WRITE_TIMEOUT"
	EnvHTTPIdleTimeout        string = "HTTP_IDLE_TIMEOUT"
	EnvHTTPShutdownTimeout    string = "HTTP_SHUTDOWN_TIMEOUT"
	EnvHTTPPrivateKeyPath     string = "HTTP_PRIVATE_KEY_PATH"
	EnvHTTPCertificatePath    string = "HTTP_CERTIFICATE_PATH"
	EnvHTTPSecure             string = "HTTP_SECURE"
	EnvHTTPMaxRequestBodySize string = "HTTP_MAX_REQUEST_BODY_SIZE"
)

// Log config envs
const (
	EnvLogLevel  string = "LOG_LEVEL"
	EnvLogFormat string = "LOG_FORMAT"
)

// redis config envs
const (
	EnvRedisHost     string = "REDIS_HOST"
	EnvRedisPort     string = "REDIS_PORT"
	EnvRedisPassword string = "REDIS_PASSWORD"
	EnvRedisDB       string = "REDIS_DB"
)

type AppEnv string

func (ae AppEnv) Exists() bool {
	return appEnvs.Contains(ae)
}

type appEnvList map[AppEnv]bool

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

var appEnvs appEnvList = map[AppEnv]bool{
	AppEnvProduction:  false,
	AppEnvDevelopment: false,
	AppEnvTest:        false,
}
