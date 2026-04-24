package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCompress_Handle(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Debug", mock.Anything, mock.Anything).Return().Maybe()

	c := NewCompress(mockLog, "application/json")

	// Создаем цепочку: Middleware -> Handler, который пишет JSON
	handler := c.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))

	t.Run("Should use Brotli when requested", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "br") // Запрашиваем brotli

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, "br", rr.Header().Get("Content-Encoding"))

		// Проверка: можно ли это реально распаковать?
		brReader := brotli.NewReader(rr.Body)
		decoded, _ := io.ReadAll(brReader)
		assert.Equal(t, `{"status":"ok"}`, string(decoded))
	})

	t.Run("Should NOT compress disallowed content type", func(t *testing.T) {
		// Создаем хендлер для картинки
		imgHandler := c.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("fake-png-data"))
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "br")

		rr := httptest.NewRecorder()
		imgHandler.ServeHTTP(rr, req)

		assert.Empty(t, rr.Header().Get("Content-Encoding"), "Should not compress png")
		assert.Equal(t, "fake-png-data", rr.Body.String())
	})
}
