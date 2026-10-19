package transport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextHelpers_FullSuite(t *testing.T) {
	baseCtx := context.Background()

	t.Run("Успешный цикл записи и чтения RequestID (Happy Path)", func(t *testing.T) {
		// Arrange
		expectedID := "req-prod-100500"

		// Act
		ctx := WithRequestID(baseCtx, expectedID)
		actualID := RequestID(ctx)

		// Assert
		assert.Equal(t, expectedID, actualID)
	})

	t.Run("Успешный цикл записи и чтения TraceID (Happy Path)", func(t *testing.T) {
		// Arrange
		expectedTrace := "trace-otel-fedcba9876543210"

		// Act
		ctx := WithTraceID(baseCtx, expectedTrace)
		actualTrace := TraceID(ctx)

		// Assert
		assert.Equal(t, expectedTrace, actualTrace)
	})

	t.Run("Успешный цикл записи и чтения RealIP (Happy Path)", func(t *testing.T) {
		// Arrange
		expectedIP := "10.0.4.15"

		// Act
		ctx := WithRealIP(baseCtx, expectedIP)
		actualIP := RealIP(ctx)

		// Assert
		assert.Equal(t, expectedIP, actualIP)
	})

	t.Run("Изоляция ключей контекста друг от друга", func(t *testing.T) {
		// Проверяем, что запись одного ключа не затирает и не пересекается с другими
		ctx := WithRequestID(baseCtx, "req-1")
		ctx = WithTraceID(ctx, "trace-2")
		ctx = WithRealIP(ctx, "127.0.0.1")

		assert.Equal(t, "req-1", RequestID(ctx))
		assert.Equal(t, "trace-2", TraceID(ctx))
		assert.Equal(t, "127.0.0.1", RealIP(ctx))
	})

	t.Run("Возврат пустой строки при отсутствии значений в контексте", func(t *testing.T) {
		emptyCtx := context.Background()

		assert.Empty(t, RequestID(emptyCtx))
		assert.Empty(t, TraceID(emptyCtx))
		assert.Empty(t, RealIP(emptyCtx))
	})

	t.Run("Безопасное гашение паник при передаче nil-контекста", func(t *testing.T) {
		// Проверяем, что благодаря utils.IsNil методы записи возвращают nil без паники
		assert.NotPanics(t, func() {
			nilWriteReq := WithRequestID(nil, "id")
			nilWriteTrc := WithTraceID(nil, "id")
			nilWriteIP := WithRealIP(nil, "ip")

			assert.Nil(t, nilWriteReq)
			assert.Nil(t, nilWriteTrc)
			assert.Nil(t, nilWriteIP)
		})

		// Проверяем, что методы чтения возвращают пустую строку на nil без паники
		assert.NotPanics(t, func() {
			assert.Empty(t, RequestID(nil))
			assert.Empty(t, TraceID(nil))
			assert.Empty(t, RealIP(nil))
		})
	})
}
