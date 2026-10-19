package domain

import (
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// TestValidateOutboxStatus реализует табличное тестирование (Table-Driven Tests)
// для верификации всех возможных состояний жизненного цикла outbox-записи.
func TestValidateOutboxStatus(t *testing.T) {
	// Описываем структуру тестового случая
	tests := []struct {
		name    string // Понятное описание сценария
		status  string // Входной аргумент для тестируемой функции
		wantErr bool   // Флаг, указывающий, ожидаем ли мы ошибку в данном кейсе
	}{
		{
			name:    "Успешная валидация статуса PENDING",
			status:  OutboxStatusPending,
			wantErr: false,
		},
		{
			name:    "Успешная валидация статуса SUCCESS",
			status:  OutboxStatusSuccess,
			wantErr: false,
		},
		{
			name:    "Успешная валидация статуса FAILED",
			status:  OutboxStatusFailed,
			wantErr: false,
		},
		{
			name:    "Сбой валидации на неизвестном текстовом статусе",
			status:  "UNKNOWN_STATUS_XYZ",
			wantErr: true,
		},
		{
			name:    "Сбой валидации на пустой строке",
			status:  "",
			wantErr: true,
		},
		{
			name:    "Сбой валидации при изменении регистра символов",
			status:  "pending", // Карта чувствительна к регистру строк
			wantErr: true,
		},
		{
			name:    "Сбой валидации при наличии фоновых пробелов",
			status:  " PENDING ",
			wantErr: true,
		},
	}

	// Запускаем итерируемый цикл по всем тест-кейсам таблицы
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutboxStatus(tt.status)

			// Проверяем факт наличия или отсутствия ошибки в соответствии с ожиданиями
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutboxStatus() ошибка = %v, ожидалось наличие ошибки: %v", err, tt.wantErr)
				return
			}

			// Если ошибка ожидалась, проводим глубокий аудит её платформенного типа
			if tt.wantErr && err != nil {
				// Проверяем, что ошибка реализует наш кастомный интерфейс BllValidateError.
				// Используем для этого стандартный идиоматичный механизм errors.As.
				if _, ok := errors.AsType[*errs.BllValidateError](err); !ok {
					t.Errorf("ValidateOutboxStatus() вернул ошибку некорректного типа; ожидался тип errs.BllValidateError, получен: %T", err)
				}
			}
		})
	}
}
