package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Кастомный маппер для изоляции тестов
func testErrorMapper(err error) int {
	if errors.Is(err, context.Canceled) {
		return http.StatusInternalServerError
	}
	return http.StatusBadRequest
}

func TestRenderEmpty(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()

	// Act
	RenderEmpty(rec, http.StatusNoContent)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestRenderJSON_HappyPath(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	testData := map[string]string{"status": "ok"}

	// Act
	RenderJSON(rec, http.StatusOK, testData, testErrorMapper)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestRenderError_Masking5xx(t *testing.T) {
	t.Run("Маскировка ошибки 500 - пустой ответ", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		serverErr := context.Canceled // Наш маппер вернет 500 на эту ошибку

		// Act
		RenderError(rec, serverErr, testErrorMapper)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Empty(t, rec.Body.String(), "Тело ответа для 5xx ошибок должно быть абсолютно пустым")
	})

	t.Run("Обычная ошибка 400 - рендеринг структуры ошибки", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		badRequestErr := errors.New("bad input data")

		// Act
		RenderError(rec, badRequestErr, testErrorMapper)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.NotEmpty(t, rec.Body.String(), "Для 4xx ошибок тело должно содержать JSON")
	})
}

func TestRenderJSON_MarshalFailure(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()

	// Передаем неподдерживаемый тип для json.Marshal (например, канал)
	unmarshalableData := make(chan int)

	// Act & Assert
	assert.NotPanics(t, func() {
		// Должен перехватить ошибку маршалинга и уйти в RenderError -> маппер вернет 400 или 500
		RenderJSON(rec, http.StatusOK, unmarshalableData, testErrorMapper)
	})
}
