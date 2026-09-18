package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

func TestKafkaLoggerImpl_InfoLogger(t *testing.T) {
	// 1. Создаем мок вашего основного логгера приложения
	mockLogger := mocks.NewMockLogger(t)

	// 2. Настраиваем ожидание: метод Infof должен быть вызван с конкретными параметрами
	expectedTemplate := "connection established to %s"
	expectedArg := "localhost:9092"

	mockLogger.On("Infof", expectedTemplate, []any{expectedArg}).Return()

	// 3. Инициализируем тестируемый объект
	kafkaLoggerImpl := logger.NewKafkaLogger(mockLogger)
	infoLogger := kafkaLoggerImpl.InfoLogger()

	// Проверяем, что вернулся не nil
	assert.NotNil(t, infoLogger)

	// 4. Вызываем метод Printf (именно так его будет вызывать библиотека kafka-go внутри)
	infoLogger.Printf(expectedTemplate, expectedArg)
}

func TestKafkaLoggerImpl_ErrorLogger(t *testing.T) {
	mockLogger := mocks.NewMockLogger(t)

	expectedTemplate := "failed to write message: %w"
	expectedArg := assert.AnError // Специальный маркер ошибки от testify

	// Настраиваем ожидание для Errorf
	mockLogger.On("Errorf", expectedTemplate, []any{expectedArg}).Return()

	kafkaLoggerImpl := logger.NewKafkaLogger(mockLogger)
	errorLogger := kafkaLoggerImpl.ErrorLogger()

	assert.NotNil(t, errorLogger)

	// Вызываем метод Printf у error-логгера
	errorLogger.Printf(expectedTemplate, expectedArg)
}
