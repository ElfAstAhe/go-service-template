package kafka

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage_NewAndExtract(t *testing.T) {
	originalKafkaMsg := kafka.Message{
		Topic: "test-topic",
		Value: []byte("test-payload"),
		Headers: []kafka.Header{
			{Key: "header-key-1", Value: []byte("header-val-1")},
			{Key: "header-key-2", Value: []byte("header-val-2")},
		},
	}

	msg := NewMessage(originalKafkaMsg)

	// Проверяем базовые геттеры интерфейса
	assert.Equal(t, "test-topic", msg.GetTargetName())
	assert.Equal(t, []byte("test-payload"), msg.GetPayload())

	props := msg.GetProperties()
	require.NotNil(t, props)
	assert.Equal(t, "header-val-1", props["header-key-1"])
	assert.Equal(t, "header-val-2", props["header-key-2"])

	// Проверяем извлечение оригинала
	rawExtracted, err := msg.ExtractOriginalMessage()
	require.NoError(t, err)

	extractedPtr, ok := rawExtracted.(*kafka.Message)
	require.True(t, ok)
	assert.Equal(t, "test-topic", extractedPtr.Topic)
	assert.Equal(t, []byte("test-payload"), extractedPtr.Value)
}

func TestExtractOriginalKafkaMessage_Helper(t *testing.T) {
	originalKafkaMsg := kafka.Message{
		Topic: "helper-topic",
		Value: []byte("helper-payload"),
	}

	msg := NewMessage(originalKafkaMsg)

	extracted, err := ExtractOriginalKafkaMessage(msg)
	require.NoError(t, err)
	assert.Equal(t, "helper-topic", extracted.Topic)
}
