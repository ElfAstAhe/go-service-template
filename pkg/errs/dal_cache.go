package errs

import (
	"fmt"
)

// DalCacheError реализует нативный interface error, специфицируя системные ошибки,
// возникающие при взаимодействии с подсистемами кэширования (Redis, In-Memory Шарды) на уровне DAL.
//
// Инкапсулирует контекст сбоя: имя операции (op), пользовательское текстовое описание (msg)
// и ссылку на исходную нижележащую причину падения рантайма кэша (err).
type DalCacheError struct {
	op  string // Имя метода репозитория или сервиса, где упал кэш (например, "UserRepository.GetFromCache")
	msg string // Человекочитаемое описание сути технического сбоя контура кэширования
	err error  // Ссылка на исходную корневую ошибку (например, ошибку отвала сокета Redis или сбой десериализации)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*DalCacheError)(nil)

// NewDalCacheError — фабричный конструктор ошибки работы с кэш-слоем DAL.
// ИСПРАВЛЕНО: Поле msg теперь корректно инициализируется в структуре и не теряется в логах.
func NewDalCacheError(op string, msg string, err error) *DalCacheError {
	return &DalCacheError{
		op:  op,
		err: err,
		msg: msg,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Последовательно склеивает контекст операции, текстовое сообщение и стек вложенных ошибок.
func (dc *DalCacheError) Error() string {
	msg := "DAL: accessing cache failed"
	if dc.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, dc.op)
	}
	if dc.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, dc.msg)
	}
	if dc.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов кэш-движка
		msg = fmt.Sprintf("%s: with error %v", msg, dc.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку кэш-движка или сетевого сокета, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As во внешних классификаторах транспорта.
func (dc *DalCacheError) Unwrap() error {
	return dc.err
}
