package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestMapToHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "Возврат 200 OK при отсутствии ошибки (nil)",
			err:  nil,
			want: http.StatusOK,
		},
		{
			name: "Возврат 400 Bad Request для InvalidArgumentError",
			err:  &errs.InvalidArgumentError{},
			want: http.StatusBadRequest,
		},
		{
			name: "Возврат 400 Bad Request для BllValidateError",
			err:  &errs.BllValidateError{},
			want: http.StatusBadRequest,
		},
		{
			name: "Возврат 401 Unauthorized для BllUnauthorizedError",
			err:  &errs.BllUnauthorizedError{},
			want: http.StatusUnauthorized,
		},
		{
			name: "Возврат 403 Forbidden для BllForbiddenError",
			err:  &errs.BllForbiddenError{},
			want: http.StatusForbidden,
		},
		{
			name: "Возврат 404 Not Found для BllNotFoundError",
			err:  &errs.BllNotFoundError{},
			want: http.StatusNotFound,
		},
		{
			name: "Возврат 404 Not Found для DalNotFoundError",
			err:  &errs.DalNotFoundError{},
			want: http.StatusNotFound,
		},
		{
			name: "Возврат 409 Conflict для BllUniqueError",
			err:  &errs.BllUniqueError{},
			want: http.StatusConflict,
		},
		{
			name: "Возврат 409 Conflict для DalAlreadyExistsError",
			err:  &errs.DalAlreadyExistsError{},
			want: http.StatusConflict,
		},
		{
			name: "Возврат 410 Gone для DalSoftDeletedError",
			err:  &errs.DalSoftDeletedError{},
			want: http.StatusGone,
		},
		{
			name: "Предохранитель: возврат 500 Internal Server Error для стандартной незнакомой ошибки",
			err:  errors.New("неизвестный сбой базы данных или сети"),
			want: http.StatusInternalServerError,
		},
		{
			name: "Предохранитель: возврат 500 Internal Server Error для обернутой неизвестной ошибки",
			err:  errors.Join(errors.New("внешний контекст"), errors.New("ошибка драйвера")),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapToHTTPStatus(tt.err)
			assert.Equal(t, tt.want, got, "Маппинг статуса для сценария '%s' завершился некорректно", tt.name)
		})
	}
}
