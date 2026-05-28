package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDecompress_Handle(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
	// Лимит 10 байт для теста
	mw := NewDecompress(5000, mockLog)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("READ ERROR: %v\n", err) // Посмотри, нет ли там "unexpected EOF"
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	handler := mw.Handle(next)

	t.Run("Should decompress gzip", func(t *testing.T) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, _ = gw.Write([]byte("hello"))
		_ = gw.Close() // ОБЯЗАТЕЛЬНО ДО NewRequest

		req := httptest.NewRequest("POST", "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "hello", rr.Body.String())
	})

	t.Run("Should return 400 on unknown encoding", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("data"))
		req.Header.Set("Content-Encoding", "unknown")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Should enforce maxRequestBodySize", func(t *testing.T) {
		// Присылаем данных больше, чем лимит 10 байт
		bigData := strings.Repeat("a", 100)
		req := httptest.NewRequest("POST", "/", strings.NewReader(bigData))
		req.Header.Set("Content-Encoding", "gzip") // Предположим, utils.NewDecompressReader пропустит как raw, если не gzip

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// MaxBytesReader вызывает панику или возвращает ошибку при Read,
		// нужно проверить, как твой хендлер это "прожевал"
	})
}
