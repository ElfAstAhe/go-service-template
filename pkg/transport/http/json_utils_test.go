package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
)

// DummyDTO используется для валидации процесса декодирования в тестах.
type DummyDTO struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		factoryDst  func() any
		wantErr     bool
		validateErr func(t *testing.T, err error)
	}{
		{
			name: "Успешный разбор валидного JSON (Happy Path)",
			body: `{"id": 42, "title": "Go 1.26 Rules"}`,
			factoryDst: func() any {
				return &DummyDTO{}
			},
			wantErr: false,
		},
		{
			name: "Ошибка: Strict mode сработал на неизвестное поле",
			body: `{"id": 42, "title": "Go", "unknown_field": "ops"}`,
			factoryDst: func() any {
				return &DummyDTO{}
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				var tlErr *errs.TlMappingError
				// Проверяем, что вернулась именно ожидаемая транспортная ошибка
				assert.ErrorAs(t, err, &tlErr)
				assert.Contains(t, err.Error(), "unknown field")
			},
		},
		{
			name: "Ошибка: синтаксически некорректный JSON (Malformed)",
			body: `{"id": 42, "title": `,
			factoryDst: func() any {
				return &DummyDTO{}
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				var tlErr *errs.TlMappingError
				assert.ErrorAs(t, err, &tlErr)
			},
		},
		{
			name: "Ошибка: передача значения вместо указателя в dst",
			body: `{"id": 1}`,
			factoryDst: func() any {
				return DummyDTO{} // Передаем структуру, а не ссылку &DummyDTO
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/api/v1/test", bytes.NewBufferString(tt.body))
			dst := tt.factoryDst()

			// Act
			err := DecodeJSON(req, dst)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.validateErr != nil {
					tt.validateErr(t, err)
				}
			} else {
				assert.NoError(t, err)
				res, ok := dst.(*DummyDTO)
				assert.True(t, ok)
				assert.Equal(t, 42, res.ID)
				assert.Equal(t, "Go 1.26 Rules", res.Title)
			}
		})
	}
}

//func TestDecodeJSON_MaxBytesLimitPanicIndicator(t *testing.T) {
//    // Исправленный тест: передаем валидный, но гигантский JSON-массив,
//    // чтобы MaxBytesReader гарантированно уперся в лимит 1 МБ при чтении потока.
//    t.Run("Fail-Fast паника при превышении лимита в 1 МБ", func(t *testing.T) {
//        // Формируем валидное начало огромного JSON-массива, превышающего 1 МБ
//        largeValidJSON := "[" + strings.Repeat("1,", 1024*1024) + "1]"
//        req := httptest.NewRequest(http.MethodPost, "/api/v1/heavy", bytes.NewBufferString(largeValidJSON))
//        var dst DummyDTO
//
//        assert.Panics(t, func() {
//            // json.Decoder начнет стримить этот массив, вычитает больше 1 МБ,
//            // после чего MaxBytesReader попытается записать ошибку в nil ResponseWriter и запаникует.
//            _ = DecodeJSON(req, &dst)
//        }, "Ожидалась паника рантайма из-за nil ResponseWriter в MaxBytesReader при чтении потока данных")
//    })
//}
