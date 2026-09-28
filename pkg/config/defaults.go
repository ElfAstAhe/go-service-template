package config

import (
	"time"
)

// FlagConfig определяет имя CLI-флага для передачи абсолютного или относительного пути к файлу конфигурации.
const FlagConfig = "config-path"

// EnvConfig определяет имя переменной окружения операционной системы для указания пути к конфигурационному файлу.
const EnvConfig string = "CONFIG_PATH"

// CLI-флаги для переопределения базовых параметров жизненного цикла приложения.
const (
	FlagAppEnv          string = "env"               // CLI-флаг режима среды окружения (prod/dev/test)
	FlagAppInitTimeout  string = "app-init-timeout"  // CLI-флаг лимита времени на сборку IoC графа зависимостей
	FlagAppStopTimeout  string = "app-stop-timeout"  // CLI-флаг лимита времени на мягкую остановку серверов и воркеров
	FlagAppCloseTimeout string = "app-close-timeout" // CLI-флаг лимита времени на принудительную деаллокацию пулов СУБД
)

// CLI-флаги подсистемы безопасности, шифрования и авторизации (IAM/Auth).
const (
	FlagAuthJWTSecret          string = "auth-jwt-secret"           // CLI-флаг секретной соли HMAC подписи JWT токенов
	FlagAuthJWTSigningMethod   string = "auth-jwt-signing-method"   // CLI-флаг алгоритма подписи JWT (например, HS256)
	FlagAuthAccessTokenTTL     string = "auth-access-token-ttl"     // CLI-флаг времени жизни Access-токена
	FlagAuthRefreshTokenTTL    string = "auth-refresh-token-ttl"    // CLI-флаг времени жизни Refresh-токена
	FlagAuthRSAPrivateKeyPath  string = "auth-rsa-private-key-path" // CLI-флаг пути к приватному RSA-ключу на диске
	FlagAuthMasterPasswordSalt string = "auth-master-password-salt" // CLI-флаг соли для хэширования паролей пользователей
)

// CLI-флаги низкоуровневых параметров подключения и тюнинга пула реляционной СУБД.
const (
	FlagDBDSN             string = "db-dsn"               // CLI-флаг строки подключения Data Source Name
	FlagDBDriver          string = "db-driver"            // CLI-флаг типа SQL-драйвера (postgres, mysql, etc.)
	FlagDBMaxOpenConns    string = "db-max-open-conns"    // CLI-флаг жесткого верхнего лимита открытых соединений СУБД
	FlagDBMaxIdleConns    string = "db-max-idle-conns"    // CLI-флаг жесткого лимита удерживаемых idle-сокетов пула
	FlagDBMaxIdleLifetime string = "db-max-idle-lifetime" // CLI-флаг лимита времени жизни idle-сессий СУБД
	FlagDBConnTimeout     string = "db-conn-timeout"      // CLI-флаг таймаута установки первичной сетевой сессии
)

// CLI-флаги для настройки транспортных параметров и KeepAlive-политик gRPC-сервера.
const (
	FlagGRPCAddress          string = "grpc-address"            // CLI-флаг сетевого адреса gRPC слушателя
	FlagGRPCMaxConnIdle      string = "grpc-max-conn-idle"      // CLI-флаг лимита простоя HTTP/2 соединений
	FlagGRPCMaxConnAge       string = "grpc-max-conn-age"       // CLI-флаг лимита времени жизни коннекта (Ротация)
	FlagGRPCMaxConnAgeGrace  string = "grpc-max-conn-age-grace" // CLI-флаг Grace-периода мягкого закрытия старых gRPC запросов
	FlagGRPCTimeout          string = "grpc-timeout"            // CLI-флаг сетевого таймаута gRPC операций
	FlagGRPCKeepAliveTime    string = "grpc-keep-alive-time"    // CLI-флаг интервала пинга KeepAlive со стороны сервера
	FlagGRPCKeepAliveTimeout string = "grpc-keep-alive-timeout" // CLI-флаг таймаута ожидания ответа на KeepAlive пинг
	FlagGRPCShutdownTimeout  string = "grpc-shutdown-timeout"   // CLI-флаг таймаута мягкого гашения gRPC сервера
)

// CLI-флаги для настройки HTTP/REST веб-сервера (размеры тела запросов, TLS и таймауты).
const (
	FlagHTTPAddress            string = "http-address"               // CLI-флаг сетевого адреса HTTP слушателя
	FlagHTTPReadTimeout        string = "http-read-timeout"          // CLI-флаг лимита времени чтения заголовков и боди запроса
	FlagHTTPWriteTimeout       string = "http-write-timeout"         // CLI-флаг лимита времени записи байт ответа клиенту
	FlagHTTPIdleTimeout        string = "http-idle-timeout"          // CLI-флаг лимита времени простоя Keep-Alive HTTP сессий
	FlagHTTPShutdownTimeout    string = "http-shutdown-timeout"      // CLI-флаг таймаута мягкого гашения HTTP сервера
	FlagHTTPPrivateKeyPath     string = "http-private-key-path"      // CLI-флаг пути к приватному ключу шифрования TLS
	FlagHTTPCertificatePath    string = "http-certificate-path"      // CLI-флаг пути к SSL/TLS сертификату веб-сервера
	FlagHTTPSecure             string = "http-secure"                // CLI-флаг активации защищенного режима HTTPS (true/false)
	FlagHTTPMaxRequestBodySize string = "http-max-request-body-size" // CLI-флаг ограничения размера боди (Защита от DOS)
)

// CLI-флаги подсистемы структурированного логирования.
const (
	FlagLogLevel  string = "log-level"  // CLI-флаг глубины логирования (debug/info/warn/error)
	FlagLogFormat string = "log-format" // CLI-флаг формата вывода (json/console)
)

// CLI-флаги подключения к распределенному кэш-хранилищу Redis.
const (
	FlagRedisHost     string = "redis-host"     // CLI-флаг сетевого хоста или IP-адреса инстанса Redis
	FlagRedisPort     string = "redis-port"     // CLI-флаг сетевого порта Redis
	FlagRedisPassword string = "redis-password" // CLI-флаг пароля авторизации
	FlagRedisDB       string = "redis-db"       // CLI-флаг индекса целевой логической базы данных Redis
)

// CLI-флаги подсистемы сквозного распределенного сбора трассировок (Telemetry / OpenTelemetry).
const (
	FlagTelemetryEnabled          string = "telemetry-enabled"           // CLI-флаг глобальной активации экспорта спанов телеметрии
	FlagTelemetryServiceName      string = "telemetry-service-name"      // CLI-флаг имени сервиса для группировки в Jaeger/OpenSearch
	FlagTelemetryExporterEndpoint string = "telemetry-exporter-endpoint" // CLI-флаг адреса OTLP коллектора телеметрии (gRPC/HTTP)
	FlagTelemetrySampleRate       string = "telemetry-sample-rate"       // CLI-флаг коэффициента семплирования трассировок (0.0 до 1.0)
	FlagTelemetryTimeout          string = "telemetry-timeout"           // CLI-флаг сетевого таймаута экспортера телеметрии
)

// Временные лимиты по умолчанию для управления жизненным циклом асинхронных раннеров (Runners).
const (
	DefaultRunnerStopTimeout  time.Duration = 15 * time.Second // Дефолтный таймаут на остановку консьюмеров или серверов
	DefaultRunnerCloseTimeout time.Duration = 5 * time.Second  // Дефолтный таймаут на уничтожение внутренних сетевых ресурсов
)

// Параметры по умолчанию для ядра жизненного цикла приложения.
const (
	DefaultAppEnv          AppEnv        = AppEnvDevelopment // Среда окружения по умолчанию (dev контур локальной разработки)
	DefaultAppInitTimeout  time.Duration = 30 * time.Second  // Дефолтный лимит на сборку и инициализацию IoC графа зависимостей
	DefaultAppStopTimeout  time.Duration = 20 * time.Second  // Дефолтный лимит на Graceful Shutdown серверов и воркеров
	DefaultAppCloseTimeout time.Duration = 5 * time.Second   // Дефолтный лимит на деаллокацию тяжелых пулов памяти и сокетов СУБД
)

// Строковые ключи Viper-маппинга для иерархического связывания внутренних параметров ядра приложения.
const (
	KeyAppEnv          string = "app.env"
	KeyAppInitTimeout  string = "app.init_timeout"
	KeyAppStopTimeout  string = "app.stop_timeout"
	KeyAppCloseTimeout string = "app.close_timeout"
)

// Константные лимиты и дефолты HTTP/REST веб-сервера.
const (
	DefaultHTTPAddress            string        = "localhost:8080" // Сетевой интерфейс и порт по умолчанию для REST-API
	DefaultHTTPSecure             bool          = false            // Отключение HTTPS по умолчанию (Plain HTTP текстовый режим)
	DefaultHTTPReadTimeout        time.Duration = 5 * time.Second  // Порог ожидания при чтении заголовков и боди входящего запроса
	DefaultHTTPWriteTimeout       time.Duration = 5 * time.Second  // Порог ожидания при передаче байт ответа обратно клиенту
	DefaultHTTPIdleTimeout        time.Duration = 30 * time.Second // Порог удержания сокета в режиме ожидания нового Keep-Alive запроса
	DefaultHTTPShutdownTimeout    time.Duration = 15 * time.Second // Лимит времени на доваривание запросов при Graceful Shutdown сервера
	DefaultHTTPMaxRequestBodySize int           = 1024 * 1024 * 4  // Верхний жесткий лимит тела HTTP запроса — 4MB (Защита от OOM)
)

// Строковые ключи Viper-маппинга для иерархического связывания параметров HTTP/REST сервера.
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
// Настройки по умолчанию и политики балансировки HTTP/2 соединений для gRPC-транспорта.
const (
	DefaultGRPCAddress string = "localhost:50051" // Сетевой интерфейс и порт по умолчанию для gRPC-слушателя

	// DefaultGRPCMaxConnIdle - Даем соединениям «отдохнуть», но не убиваем их сразу, снижая нагрузку на сокеты.
	DefaultGRPCMaxConnIdle time.Duration = 5 * time.Minute

	// DefaultGRPCMaxConnAge - Ротируем соединения раз в 20 минут для равномерной балансировки трафика внутри K8s.
	DefaultGRPCMaxConnAge time.Duration = 20 * time.Minute

	// DefaultGRPCMaxConnAgeGrace - Grace-период, чтобы активные gRPC-стримы успели довариться при ротации коннекта.
	DefaultGRPCMaxConnAgeGrace  time.Duration = 1 * time.Minute
	DefaultGRPCTimeout          time.Duration = 5 * time.Second     // Жесткий порог ожидания выполнения gRPC операции
	DefaultGRPCKeepAliveTime    time.Duration = 2 * 2 * time.Minute // Интервал превентивной отправки пингов для проверки живой TCP-сессии
	DefaultGRPCKeepAliveTimeout time.Duration = 20 * time.Second    // Время ожидания ответа на пинг до принудительного дропа сокета
	DefaultGRPCShutdownTimeout  time.Duration = 15 * time.Second    // Лимит времени на мягкую остановку gRPC обработчиков
)

// ====================================================================
// gRPC
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания параметров gRPC сервера.
const (
	KeyGRPCAddress          string = "grpc.address"
	KeyGRPCMaxConnIdle      string = "grpc.max-conn-idle"
	KeyGRPCMaxConnAge       string = "grpc.max-conn-age"
	KeyGRPCMaxConnAgeGrace  string = "grpc-max-conn-age-grace"
	KeyGRPCTimeout          string = "grpc.timeout"
	KeyGRPCKeepAliveTime    string = "grpc-keep-alive-time"
	KeyGRPCKeepAliveTimeout string = "grpc-keep-alive-timeout"
	KeyGRPCShutdownTimeout  string = "grpc.shutdown_timeout"
)

// ====================================================================
// logger defaults
// ====================================================================
// Настройки по умолчанию для подсистемы логирования.
const (
	DefaultLogLevel  string = "info"    // Базовый уровень фильтрации логов фреймворка
	DefaultLogFormat string = "console" // Базовый формат вывода (console для локальной разработки)
)

// ====================================================================
// logging
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания параметров логирования.
const (
	KeyLogLevel  string = "log.level"
	KeyLogFormat string = "log.format"
)

// ====================================================================
// DB defaults (only pool settings)
// ====================================================================
// Конфигурация по умолчанию для пула соединений реляционной базы данных (Prod-Ready конфигурация).
const (
	DefaultDBDriver              string        = ""               // Инициализируется пустым, требуя явного переопределения (postgres/mysql)
	DefaultDBDSN                 string        = ""               // Строка подключения (намеренно скрыта для безопасности)
	DefaultDBMaxOpenConns        int           = 32               // Верхний лимит одновременно открытых сетевых сокетов к СУБД
	DefaultDBMaxIdleConns        int           = 4                // Жесткий лимит удерживаемых простаивающих сокетов в пуле
	DefaultDBConnMaxIdleLifetime time.Duration = 60 * time.Second // Лимит жизненного цикла idle-сессий для защиты от закрытия балансировщиком
	DefaultDBConnTimeout         time.Duration = 30 * time.Second // Порог ожидания при первичной установке TCP-сессии с базой
)

// ====================================================================
// DB
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания параметров реляционной СУБД.
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
// Настройки по умолчанию для подсистемы сквозной распределенной трассировки OpenTelemetry.
const (
	DefaultTelemetryEnabled          bool          = false            // По умолчанию сбор спанов отключен во избежание оверхеда в dev-среде
	DefaultTelemetryExporterEndpoint string        = "localhost:4317" // Стандартный gRPC порт OTLP коллектора (Jaeger/OpenSearch)
	DefaultTelemetrySampleRate       float64       = 1.0              // Коэффициент семплирования 1.0 означает сбор 100% трассировок
	DefaultTelemetryTimeout          time.Duration = 5 * time.Second  // Порог ожидания при отправке батча спанов в коллектор
)

// ====================================================================
// telemetry
// ====================================================================
// Строковые ключи Viper-маппинга для подсистемы сбора OpenTelemetry распределенных трассировок.
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
// Константные таймауты авторизации и параметры безопасности сессий по умолчанию.
const (
	DefaultAuthSigningMethod   string        = "HS256"          // Метод подписи токенов по умолчанию: HMAC-SHA256
	DefaultAuthAccessTokenTTL  time.Duration = 15 * time.Minute // Время жизни короткоживущего токена доступа (Access Token)
	DefaultAuthRefreshTokenTTL time.Duration = 24 * time.Hour   // Время жизни сессионного токена обновления (Refresh Token)
)

// ====================================================================
// Auth
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания секретов подсистемы безопасности и IAM.
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
// Настройки по умолчанию для сетевого пула соединений брокеров сообщений по протоколу AMQP 0-9-1.
const (
	DefaultAMQPConnectorURL             string        = "amqp://localhost:5672/" // Сетевой адрес брокера RabbitMQ по умолчанию
	DefaultAMQPConnectorConnectTimeout  time.Duration = 10 * time.Second         // Лимит времени на установку TCP-сессии
	DefaultAMQPConnectorWriteTimeout    time.Duration = 10 * time.Second         // Лимит времени на запись фрейма в брокер
	DefaultAMQPConnectorIdleTimeout     time.Duration = 30 * time.Second         // Лимит удержания простаивающего канала со стороны ОС
	DefaultAMQPConnectorShutdownTimeout time.Duration = 15 * time.Second         // Лимит времени на мягкое освобождение ресурсов канала
)

// ====================================================================
// amqp sender
// ====================================================================
// Параметры по умолчанию политики повторных попыток отправки фреймов (Publish Retry Policy) продюсера.
const (
	DefaultAMQPSenderConnectTimeout        time.Duration = 10 * time.Second       // Таймаут логического открытия продюсер-канала
	DefaultAMQPSenderShutdownTimeout       time.Duration = 15 * time.Second       // Таймаут на принудительный Flush буферов отправки
	DefaultAMQPSenderPublishMaxTryAttempts int           = 2                      // Лимит попыток публикации до возвращения критической ошибки
	DefaultAMQPSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond // Стартовая задержка экспоненциального бэккоффа ретраев
	DefaultAMQPSenderPublishMaxRetryDelay  time.Duration = 3 * time.Second        // Верхний жесткий потолок ожидания между попытками
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
// Системные параметры по умолчанию для высокопроизводительного асинхронного продюсера Кафки.
const (
	// DefaultKafkaSenderConnectTimeout задает лимит времени на установку сетевого соединения с брокерами.
	DefaultKafkaSenderConnectTimeout time.Duration = 10 * time.Second

	// DefaultKafkaSenderIdleTimeout задаёт время простоя сетевых соединений в пуле коннектов.
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
// Системные параметры по умолчанию для распределенного отказоустойчивого консьюмера Кафки.
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

// Глобальные переменные-синглтоны для инициализации сетевой топологии брокеров очередей.
var (
	// DefaultKafkaBrokers хранит срез хостов брокеров кластера по умолчанию для локальной разработки.
	DefaultKafkaBrokers = []string{"localhost:9092"}
)
