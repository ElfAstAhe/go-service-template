package config

const (
	// DefaultConfigFile default config file path
	DefaultConfigFile = "config/config.yaml"
)

// Брокеры kafka
const (
	// DefaultKafkaBroker kafka broker по умолчанию
	DefaultKafkaBroker string = "localhost:9092"
)

// Глобальные переменные-синглтоны для инициализации сетевой топологии брокеров очередей.
var (
	// DefaultKafkaBrokers хранит срез хостов брокеров кластера по умолчанию для локальной разработки.
	DefaultKafkaBrokers = []string{DefaultKafkaBroker}
)
