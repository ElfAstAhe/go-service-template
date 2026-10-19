package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestGetQueryInt_And_Default(t *testing.T) {
	t.Run("Успешное получение int", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?age=26", nil)
		val, err := GetQueryInt(req, "age")
		assert.NoError(t, err)
		assert.Equal(t, 26, val)
	})

	t.Run("Ошибка парсинга нечислового значения", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?age=abc", nil)
		_, err := GetQueryInt(req, "age")
		var invErr *errs.InvalidArgumentError
		assert.ErrorAs(t, err, &invErr)
	})

	t.Run("Возврат дефолта при отсутствии значения", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		val := GetQueryIntDefault(req, "age", 18)
		assert.Equal(t, 18, val)
	})
}

func TestGetQueryBool_And_Default(t *testing.T) {
	t.Run("Успешное получение bool", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?active=true", nil)
		val, err := GetQueryBool(req, "active")
		assert.NoError(t, err)
		assert.True(t, val)
	})

	t.Run("Возврат дефолта при ошибке синтаксиса", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?active=not-a-bool", nil)
		val := GetQueryBoolDefault(req, "active", true)
		assert.True(t, val)
	})
}

func TestGetQueryTime(t *testing.T) {
	t.Run("Успешный парсинг RFC3339", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?created_at=2026-10-20T00:30:00Z", nil)
		val, err := GetQueryTime(req, "created_at")
		assert.NoError(t, err)
		assert.Equal(t, 2026, val.Year())
	})

	t.Run("Ошибка парсинга кривого формата времени", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?created_at=20-10-2026", nil)
		_, err := GetQueryTime(req, "created_at")
		assert.Error(t, err)
	})
}

func TestGetQueryIntArray_And_Default(t *testing.T) {
	t.Run("Успешный парсинг массива чисел через запятую", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?ids=1,2,3,4", nil)
		val, err := GetQueryIntArray(req, "ids")
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4}, val)
	})

	t.Run("Ошибка при наличии букв в массиве", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?ids=1,two,3", nil)
		_, err := GetQueryIntArray(req, "ids")
		assert.Error(t, err)
	})

	t.Run("Возврат defaultValue при ошибке", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?ids=bad", nil)
		def := []int{9, 9}
		val := GetQueryIntArrayDefault(req, "ids", def)
		assert.Equal(t, def, val)
	})
}
