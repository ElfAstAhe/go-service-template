package errs

import (
	"fmt"
)

// DalError реализует нативный интерфейс error, представляя собой базовую системную ошибку
// слоя доступа к данным (Data Access Layer Error).
//
// Инкапсулирует контекст сбоя: имя низкоуровневой SQL-операции (op), пользовательское текстовое
// описание проблемы (msg) и ссылку на исходную корневую причину падения транзакции СУБД (err).
type DalError struct {
	op  string // Имя метода репозитория или хелпера, где произошел сбой (например, "Helper.Get")
	msg string // Человекочитаемое прикладное описание сути технического сбоя баз данных
	err error  // Ссылка на исходную ошибку драйвера СУБД (например, pq.Error или sql.ErrConnDone)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*DalError)(nil)

// NewDalError — фабричный конструктор базовой ошибки слоя доступа к данным (DAL).
func NewDalError(op, msg string, err error) *DalError {
	return &DalError{op: op, msg: msg, err: err}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически выстраивает текстовое сообщение, изолируя операцию квадратными скобками для читаемости в логах.
func (de *DalError) Error() string {
	msg := "DAL: error"
	if de.op != "" {
		msg = fmt.Sprintf("%s: [%s]", msg, de.op)
	}
	if de.msg != "" {
		msg = fmt.Sprintf("%s %s", msg, de.msg)
	}
	if de.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов драйвера
		msg = fmt.Sprintf("%s: %v", msg, de.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку драйвера или пула соединений, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As во внешних классификаторах транспорта.
func (de *DalError) Unwrap() error {
	return de.err
}
