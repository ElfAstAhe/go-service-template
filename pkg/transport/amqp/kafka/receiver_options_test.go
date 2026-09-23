package kafka

import (
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

// TestReceiverOptions_Validate проверяет граничные условия рантайм-валидатора опций.
func TestReceiverOptions_Validate(t *testing.T) {
	// Базовая фабрика на 100% валидных рантайм-опций в режиме группы
	baseOpts := func() *ReceiverOptions {
		return &ReceiverOptions{
			Brokers:           []string{"localhost:9092"},
			TargetName:        "tiny.auth.events",
			GroupID:           "auth-group",
			Partition:         -1, // Дефолт (режим группы)
			ConnectTimeout:    5 * time.Second,
			ShutdownTimeout:   15 * time.Second,
			MinBytes:          1024,
			MaxBytes:          1048576,
			MaxWait:           500 * time.Millisecond,
			HeartbeatInterval: 3 * time.Second,
			SessionTimeout:    30 * time.Second,
			RebalanceTimeout:  60 * time.Second,
			ReadBatchTimeout:  10 * time.Second,
			MaxAttempts:       3,
			QueueCapacity:     100,
			StartOffset:       "last",
			Logger:            mocks.NewMockLogger(t),
		}
	}

	// Сценарий 1: Успех в режиме Consumer Group
	t.Run("success_group_mode", func(t *testing.T) {
		opts := baseOpts()
		assert.NoError(t, opts.Validate())
	})

	// Сценарий 2: Успех в Standalone режиме (Direct Partition)
	t.Run("success_standalone_mode", func(t *testing.T) {
		opts := baseOpts()
		opts.GroupID = ""  // Сбрасываем группу
		opts.Partition = 0 // Явно указываем 0-ю партицию
		assert.NoError(t, opts.Validate())
	})

	// Сценарии ошибок (XOR и тайминги сокетов)
	tests := []struct {
		name           string
		mutate         func(o *ReceiverOptions)
		expectedPhrase string
	}{
		{
			name: "missing_both_group_and_partition",
			mutate: func(o *ReceiverOptions) {
				o.GroupID = ""
				o.Partition = -1
			},
			expectedPhrase: "either GroupID must be set or Partition must be >= 0",
		},
		{
			name: "mutually_exclusive_violation",
			mutate: func(o *ReceiverOptions) {
				o.Partition = 2 // И группа заполнена, и партиция указана одновременно
			},
			expectedPhrase: "mutually exclusive options",
		},
		{
			name: "read_batch_timeout_too_short",
			mutate: func(o *ReceiverOptions) {
				o.ReadBatchTimeout = 200 * time.Millisecond // 200ms <= MaxWait(500ms) -> Ложный EOF
			},
			expectedPhrase: "strictly greater than MaxWait",
		},
		{
			name: "invalid_start_offset_string",
			mutate: func(o *ReceiverOptions) {
				o.StartOffset = "from-yesterday"
			},
			expectedPhrase: "allowed values are strictly 'first' or 'last'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := baseOpts()
			tt.mutate(opts)

			err := opts.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedPhrase)
		})
	}
}
