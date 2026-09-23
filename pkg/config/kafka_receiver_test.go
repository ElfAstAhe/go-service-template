package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 1. Тест успешной валидации в режиме Consumer Group (Happy Path 1)
func TestKafkaReceiverConfig_Validate_Group_Success(t *testing.T) {
	cfg := NewKafkaReceiverConfig(
		[]string{"localhost:9092"},
		"tiny.auth.events",
		"tiny.audit.group", // Заполнена группа
		-1,
		5*time.Second,
		15*time.Second,
		1024,
		1048576,
		500*time.Millisecond,
		"", "",
		3*time.Second,
		30*time.Second,
		60*time.Second,
		10*time.Second,
		3,
		100,
		"last",
	)

	err := cfg.Validate()
	assert.NoError(t, err)
}

// 2. Тест успешной валидации в Standalone-режиме (Happy Path 2)
func TestKafkaReceiverConfig_Validate_Standalone_Success(t *testing.T) {
	cfg := NewKafkaReceiverConfig(
		[]string{"localhost:9092"},
		"tiny.auth.events",
		"", // ИСПРАВЛЕНО: Группы нет (пустая строка)
		0,
		5*time.Second,
		15*time.Second,
		1024,
		1048576,
		500*time.Millisecond,
		"", "",
		3*time.Second,
		30*time.Second,
		60*time.Second,
		10*time.Second,
		3,
		100,
		"last",
	)

	err := cfg.Validate()
	assert.NoError(t, err)
}

// 3. Табличный тест на негативные сценарии (Edge Cases)
func TestKafkaReceiverConfig_Validate_Failures(t *testing.T) {
	// Базовый рабочий конфиг в режиме группы
	baseCfg := func() *KafkaReceiverConfig {
		c := NewKafkaReceiverConfig(
			[]string{"localhost:9092"},
			"tiny.auth.events",
			"tiny.audit.group",
			-1,
			5*time.Second,
			15*time.Second,
			1024,
			1048576,
			500*time.Millisecond,
			"", "",
			3*time.Second,
			30*time.Second,
			60*time.Second,
			10*time.Second,
			3,
			100,
			"last",
		)
		return c
	}

	tests := []struct {
		name           string
		mutate         func(c *KafkaReceiverConfig)
		expectedPhrase string
	}{
		{
			name: "empty_brokers_list",
			mutate: func(c *KafkaReceiverConfig) {
				c.Brokers = []string{}
			},
			expectedPhrase: "at least one broker address is required",
		},
		{
			name: "missing_both_group_and_partition",
			mutate: func(c *KafkaReceiverConfig) {
				c.GroupID = ""
				c.Partition = -1 // Обе сущности не заполнены
			},
			expectedPhrase: "either GroupID must be set or Partition must be >= 0",
		},
		{
			name: "mutually_exclusive_violation_both_filled",
			mutate: func(c *KafkaReceiverConfig) {
				c.Partition = 0 // И группа заполнена, и партиция указана
			},
			expectedPhrase: "mutually exclusive options",
		},
		{
			name: "gold_rule_violation_short_session",
			mutate: func(c *KafkaReceiverConfig) {
				c.HeartbeatInterval = 5 * time.Second
				c.SessionTimeout = 10 * time.Second
			},
			expectedPhrase: "SessionTimeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseCfg()
			tt.mutate(cfg)

			err := cfg.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedPhrase)
		})
	}
}
