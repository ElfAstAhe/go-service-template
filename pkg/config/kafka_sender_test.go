package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 1. Тест успешной валидации корректного конфига (Happy Path)
func TestKafkaSenderConfig_Validate_Success(t *testing.T) {
	cfg := NewKafkaSenderConfig(
		[]string{"localhost:9092"},
		"tiny.auth.events",
		5*time.Second,
		10*time.Second,
		3,
		100*time.Millisecond,
		2*time.Second,
		"", "", // Без авторизации для локалхоста
		100,
		1048576,
		10*time.Millisecond,
		5*time.Second,
		-1, // RequiredAcks All
	)

	err := cfg.Validate()
	assert.NoError(t, err)
}

// 2. Тест создания дефолтного конфига
func TestKafkaSenderConfig_NewDefault(t *testing.T) {
	cfg := NewDefaultKafkaSenderConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, DefaultKafkaBrokers, cfg.Brokers)
}

// 3. Табличный тест на все негативные сценарии (Edge Cases)
func TestKafkaSenderConfig_Validate_Failures(t *testing.T) {
	// Фабрика базового валидного конфига для изоляции подтестов
	baseCfg := func() *KafkaSenderConfig {
		return NewKafkaSenderConfig(
			[]string{"localhost:9092"},
			"tiny.auth.events",
			5*time.Second,
			10*time.Second,
			3,
			100*time.Millisecond,
			2*time.Second,
			"", "",
			100,
			1048576,
			10*time.Millisecond,
			5*time.Second,
			-1,
		)
	}

	tests := []struct {
		name           string
		mutate         func(c *KafkaSenderConfig)
		expectedPhrase string // Подстрока, которую мы гарантированно ищем в тексте ошибки
	}{
		{
			name: "empty_brokers_list",
			mutate: func(c *KafkaSenderConfig) {
				c.Brokers = []string{}
			},
			expectedPhrase: "at least one broker address is required",
		},
		{
			name: "empty_target_name_topic",
			mutate: func(c *KafkaSenderConfig) {
				c.TargetName = "   "
			},
			expectedPhrase: "empty",
		},
		{
			name: "invalid_connect_timeout",
			mutate: func(c *KafkaSenderConfig) {
				c.ConnectTimeout = 0
			},
			expectedPhrase: "ConnectTimeout",
		},
		{
			name: "invalid_shutdown_timeout",
			mutate: func(c *KafkaSenderConfig) {
				c.ShutdownTimeout = -5 * time.Second
			},
			expectedPhrase: "ShutdownTimeout",
		},
		{
			name: "zero_publish_attempts",
			mutate: func(c *KafkaSenderConfig) {
				c.PublishMaxTryAttempts = 0
			},
			expectedPhrase: "PublishMaxTryAttempts",
		},
		{
			name: "base_retry_delay_greater_than_max",
			mutate: func(c *KafkaSenderConfig) {
				c.PublishBaseRetryDelay = 5 * time.Second
				c.PublishMaxRetryDelay = 2 * time.Second
			},
			expectedPhrase: "less than base delay",
		},
		{
			name: "invalid_batch_size",
			mutate: func(c *KafkaSenderConfig) {
				c.BatchSize = 0
			},
			expectedPhrase: "BatchSize/BatchBytes",
		},
		{
			name: "invalid_batch_bytes",
			mutate: func(c *KafkaSenderConfig) {
				c.BatchBytes = -10
			},
			expectedPhrase: "BatchSize/BatchBytes",
		},
		{
			name: "required_acks_too_low",
			mutate: func(c *KafkaSenderConfig) {
				c.RequiredAcks = -2
			},
			expectedPhrase: "RequiredAcks",
		},
		{
			name: "required_acks_too_high",
			mutate: func(c *KafkaSenderConfig) {
				c.RequiredAcks = 2
			},
			expectedPhrase: "RequiredAcks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseCfg()
			tt.mutate(cfg)

			err := cfg.Validate()

			// Проверяем факт наличия ошибки через testify/assert
			assert.Error(t, err)
			// Проверяем, что в тексте ошибки четко написано, какое поле её вызвало
			assert.Contains(t, err.Error(), tt.expectedPhrase)
		})
	}
}
