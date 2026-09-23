package config

import (
	"time"
)

// FlagConfig - файл конфигурации
const FlagConfig = "config-path"

// EnvConfig - файл конфигурации
const EnvConfig string = "CONFIG_PATH"

// App config flags
const (
	FlagAppEnv          string = "env"
	FlagAppInitTimeout  string = "app-init-timeout"
	FlagAppStopTimeout  string = "app-stop-timeout"
	FlagAppCloseTimeout string = "app-close-timeout"
)

// Auth config flags
const (
	FlagAuthJWTSecret          string = "auth-jwt-secret"
	FlagAuthJWTSigningMethod   string = "auth-jwt-signing-method"
	FlagAuthAccessTokenTTL     string = "auth-access-token-ttl"
	FlagAuthRefreshTokenTTL    string = "auth-refresh-token-ttl"
	FlagAuthRSAPrivateKeyPath  string = "auth-rsa-private-key-path"
	FlagAuthMasterPasswordSalt string = "auth-master-password-salt"
)

// DB config flags
const (
	FlagDBDSN             string = "db-dsn"
	FlagDBDriver          string = "db-driver"
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
	FlagGRPCMaxConnAgeGrace  string = "grpc-max-conn-age-grace"
	FlagGRPCTimeout          string = "grpc-timeout"
	FlagGRPCKeepAliveTime    string = "grpc-keep-alive-time"
	FlagGRPCKeepAliveTimeout string = "grpc-keep-alive-timeout"
	FlagGRPCShutdownTimeout  string = "grpc-shutdown-timeout"
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

// ====================================================================
// telemetry
// ====================================================================
const (
	FlagTelemetryEnabled          string = "telemetry-enabled"
	FlagTelemetryServiceName      string = "telemetry-service-name"
	FlagTelemetryExporterEndpoint string = "telemetry-exporter-endpoint"
	FlagTelemetrySampleRate       string = "telemetry-sample-rate"
	FlagTelemetryTimeout          string = "telemetry-timeout"
)

// ====================================================================
// runner defaults
// ====================================================================
const (
	DefaultRunnerStopTimeout  time.Duration = 15 * time.Second
	DefaultRunnerCloseTimeout time.Duration = 5 * time.Second
)

// ====================================================================
// App defaults
// ====================================================================
const (
	DefaultAppEnv          AppEnv        = AppEnvDevelopment
	DefaultAppInitTimeout  time.Duration = 30 * time.Second
	DefaultAppStopTimeout  time.Duration = 20 * time.Second
	DefaultAppCloseTimeout time.Duration = 5 * time.Second
)

// ====================================================================
// App
// ====================================================================
const (
	KeyAppEnv          string = "app.env"
	KeyAppInitTimeout  string = "app.init_timeout"
	KeyAppStopTimeout  string = "app.stop_timeout"
	KeyAppCloseTimeout string = "app.close_timeout"
)

// ====================================================================
// HTTP defaults
// ====================================================================
const (
	DefaultHTTPAddress            string        = "localhost:8080"
	DefaultHTTPSecure             bool          = false
	DefaultHTTPReadTimeout        time.Duration = 5 * time.Second
	DefaultHTTPWriteTimeout       time.Duration = 5 * time.Second
	DefaultHTTPIdleTimeout        time.Duration = 30 * time.Second
	DefaultHTTPShutdownTimeout    time.Duration = 15 * time.Second
	DefaultHTTPMaxRequestBodySize int           = 1024 * 1024 * 4
)

// ====================================================================
// http
// ====================================================================
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

// ====================================================================
// gRPC defaults
// ====================================================================
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

// ====================================================================
// gRPC
// ====================================================================
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

// ====================================================================
// logger defaults
// ====================================================================
const (
	DefaultLogLevel  string = "info"
	DefaultLogFormat string = "console"
)

// ====================================================================
// logging
// ====================================================================
const (
	KeyLogLevel  string = "log.level"
	KeyLogFormat string = "log.format"
)

// ====================================================================
// DB defaults (only pool settings)
// ====================================================================
const (
	DefaultDBDriver              string        = ""
	DefaultDBDSN                 string        = ""
	DefaultDBMaxOpenConns        int           = 32
	DefaultDBMaxIdleConns        int           = 4
	DefaultDBConnMaxIdleLifetime time.Duration = 60 * time.Second
	DefaultDBConnTimeout         time.Duration = 30 * time.Second
)

// ====================================================================
// DB
// ====================================================================
const (
	KeyDBDriver              string = "db.driver"
	KeyDBDSN                 string = "db.dsn"
	KeyDBMaxOpenConns        string = "db.max_open_conns"
	KeyDBMaxIdleConns        string = "db.max_idle_conns"
	KeyDBConnMaxIdleLifetime string = "db.conn_max_idle_lifetime"
	KeyDBConnTimeout         string = "db.conn_timeout"
)

// ====================================================================
// Telemetry defaults
// ====================================================================
const (
	DefaultTelemetryEnabled          bool          = false
	DefaultTelemetryExporterEndpoint string        = "localhost:4317"
	DefaultTelemetrySampleRate       float64       = 1.0
	DefaultTelemetryTimeout          time.Duration = 5 * time.Second
)

// ====================================================================
// telemetry
// ====================================================================
const (
	KeyTelemetryEnabled          string = "telemetry.enabled"
	KeyTelemetryServiceName      string = "telemetry.service_name"
	KeyTelemetryExporterEndpoint string = "telemetry.exporter_endpoint"
	KeyTelemetrySampleRate       string = "telemetry.sample_rate"
	KeyTelemetryTimeout          string = "telemetry.timeout"
)

// ====================================================================
// access defaults
// ====================================================================
const (
	DefaultAuthSigningMethod   string        = "HS256"
	DefaultAuthAccessTokenTTL  time.Duration = 15 * time.Minute
	DefaultAuthRefreshTokenTTL time.Duration = 24 * time.Hour
)

// ====================================================================
// Auth
// ====================================================================
const (
	KeyAuthJWTSecret          string = "auth.jwt_secret"
	KeyAuthJWTSigningMethod   string = "auth.jwt_signing_method"
	KeyAuthAccessTokenTTL     string = "auth.access_token_ttl"
	KeyAuthRefreshTokenTTL    string = "auth.refresh_token_ttl"
	KeyAuthRSAPrivateKeyPath  string = "auth.rsa_private_key_path"
	KeyAuthMasterPasswordSalt string = "auth.master_password_salt"
)

// ====================================================================
// amqp connector
// ====================================================================
const (
	DefaultAMQPConnectorURL             string        = "amqp://localhost:5672/"
	DefaultAMQPConnectorConnectTimeout  time.Duration = 10 * time.Second
	DefaultAMQPConnectorWriteTimeout    time.Duration = 10 * time.Second
	DefaultAMQPConnectorIdleTimeout     time.Duration = 30 * time.Second
	DefaultAMQPConnectorShutdownTimeout time.Duration = 15 * time.Second
)

// ====================================================================
// amqp sender
// ====================================================================
const (
	DefaultAMQPSenderConnectTimeout        time.Duration = 10 * time.Second
	DefaultAMQPSenderShutdownTimeout       time.Duration = 15 * time.Second
	DefaultAMQPSenderPublishMaxTryAttempts int           = 2
	DefaultAMQPSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond
	DefaultAMQPSenderPublishMaxRetryDelay  time.Duration = 3 * time.Second
)

// ====================================================================
// amqp receiver
// ====================================================================
const (
	DefaultAMQPReceiverConnectTimeout  time.Duration = 10 * time.Second
	DefaultAMQPReceiverShutdownTimeout time.Duration = 15 * time.Second
	DefaultAMQPReceiverPrefetchCredit  int           = 100
)

// ====================================================================
// Kafka Sender (Producer) Defaults
// ====================================================================
const (
	// DefaultKafkaSenderConnectTimeout задает лимит времени на установку сетевого соединения с брокерами.
	DefaultKafkaSenderConnectTimeout time.Duration = 10 * time.Second

	// DefaultKafkaSenderIdleTimeout задаёт время простоя
	DefaultKafkaSenderIdleTimeout time.Duration = 60 * time.Second

	// DefaultKafkaSenderShutdownTimeout определяет время, выделяемое врайтеру на плавное закрытие.
	// 15 секунд гарантируют успешный сброс (flushing) асинхронных буферов из памяти на диски брокеров при остановке пода.
	DefaultKafkaSenderShutdownTimeout time.Duration = 15 * time.Second

	// DefaultKafkaSenderPublishMaxTryAttempts — количество попыток публикации сообщения при сетевых сбоях (Network Flaps).
	// Повторами на транспортном уровне мы управляем сами в методе Publish.
	DefaultKafkaSenderPublishMaxTryAttempts int = 2

	// DefaultKafkaSenderPublishBaseRetryDelay — начальная задержка перед первой повторной отправкой для экспоненциального бэкоффа.
	DefaultKafkaSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond

	// DefaultKafkaSenderPublishMaxRetryDelay — жесткий верхний лимит задержки между повторными попытками отправки.
	DefaultKafkaSenderPublishMaxRetryDelay time.Duration = 3 * time.Second

	// ПОЛЯ ПРОИЗВОДИТЕЛЬНОСТИ (Асинхронный батчинг)

	// DefaultKafkaSenderBatchSize — лимит количества сообщений в локальном буфере перед отправкой пачки.
	DefaultKafkaSenderBatchSize int = 100

	// DefaultKafkaSenderBatchBytes — максимальный объем одной пачки в байтах (1 MB).
	DefaultKafkaSenderBatchBytes int = 1048576

	// DefaultKafkaSenderBatchTimeout — 10ms. Время ожидания накопления пачки.
	// Исключает секундную задержку библиотеки по умолчанию, обеспечивая минимальный latency.
	DefaultKafkaSenderBatchTimeout time.Duration = 10 * time.Millisecond

	// DefaultKafkaSenderWriteTimeout — жесткий таймаут сокета на операцию отправки пачки в сеть.
	DefaultKafkaSenderWriteTimeout time.Duration = 10 * time.Second

	// DefaultKafkaSenderRequiredAcks — уровень подтверждения записи брокером (-1 означает "all" — весь ISR пул).
	DefaultKafkaSenderRequiredAcks int = -1
)

// ====================================================================
// Kafka Receiver (Consumer) Defaults
// ====================================================================
const (
	// DefaultKafkaReceiverConnectTimeout задает лимит времени на подключение к координатору группы брокеров.
	DefaultKafkaReceiverConnectTimeout time.Duration = 10 * time.Second

	// DefaultKafkaReceiverShutdownTimeout — время на безопасную остановку чтения и корректный выход из Consumer Group.
	DefaultKafkaReceiverShutdownTimeout time.Duration = 15 * time.Second

	// DefaultKafkaReceiverMinBytes — минимальный объем данных (в байтах), который брокер должен собрать перед ответом.
	// Снижено до 1024 (1 KB) для dev/test окружений, чтобы одиночные мелкие сообщения доставлялись без задержек.
	DefaultKafkaReceiverMinBytes int = 1024

	// DefaultKafkaReceiverMaxBytes — максимальный объем данных, принимаемый за одну сетевую трансляцию (итерацию Fetch).
	// Значение 10e6 (10 MB) защищает consumer-группу от застревания при обработке «жирных» JSON-пакетов.
	DefaultKafkaReceiverMaxBytes int = 10e6

	// DefaultKafkaReceiverMaxWait — максимальное время ожидания брокера, если объем данных еще не достиг лимита MinBytes.
	// Значение 500ms на dev/test экономит ресурсы CPU сервера, предотвращая «горячий цикл» пустых запросов.
	DefaultKafkaReceiverMaxWait time.Duration = 500 * time.Millisecond

	// DefaultKafkaReceiverHeartbeatInterval — частота фонового пинга ("я жив") к координатору группы.
	DefaultKafkaReceiverHeartbeatInterval time.Duration = 3 * time.Second

	// DefaultKafkaReceiverSessionTimeout — таймаут отсутствия пингов, после которого брокер считает под мертвым и запускает ребаланс.
	DefaultKafkaReceiverSessionTimeout time.Duration = 30 * time.Second

	// DefaultKafkaReceiverRebalanceTimeout — время, выделяемое воркеру на доработку текущей пачки и сдачу партиций при ребалансе.
	DefaultKafkaReceiverRebalanceTimeout time.Duration = 60 * time.Second

	// DefaultKafkaReceiverReadTimeout — низкоуровневый таймаут сетевого сокета на чтение. Уходит в Dialer.
	DefaultKafkaReceiverReadTimeout time.Duration = 10 * time.Second

	// DefaultKafkaReceiverMaxAttempts — количество попыток переподключения до возврата критической ошибки.
	DefaultKafkaReceiverMaxAttempts int = 3

	// DefaultKafkaReceiverQueueCapacity — емкость внутреннего фонового буфера предвыборки сообщений из сети.
	DefaultKafkaReceiverQueueCapacity int = 100

	// DefaultKafkaReceiverStartOffset определяет точку старта, если для Consumer Group еще нет сохраненного оффсета ("first" или "last").
	DefaultKafkaReceiverStartOffset string = "first"
)

var (
	// DefaultKafkaBrokers хранит срез хостов брокеров кластера по умолчанию для локальной разработки.
	DefaultKafkaBrokers = []string{"localhost:9092"}
)
