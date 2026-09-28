package container

import (
	"context"
)

// SimpleCloser описывает каноничный контракт синхронного закрытия ресурсов стандартной библиотеки Go.
// Применяется для файлов, логгеров и базовых потоков ввода-вывода.
type SimpleCloser interface {
	Close() error
}

// ContextCloser описывает современный контракт асинхронного контролируемого закрытия ресурсов (Cloud-Native Cleanup).
// Протаскивание context.Context позволяет жестко лимитировать время гашения пулов (PostgreSQL, Redis, Kafka) во время деплоя.
type ContextCloser interface {
	Close(ctx context.Context) error
}

// closeInstance пакует в себя метаданные зарегистрированного компонента,
// требующего обязательного вызова деструктора при уничтожении DI-контейнера.
type closeInstance struct {
	Name     string // Уникальное строковое имя компонента для детализации системных логов очистки
	Instance any    // Ссылка на нетипизированный физический объект зависимости (any)
}

// newCloseInstance — фабричный конструктор служебной мета-структуры закрытия ресурсов.
func newCloseInstance(name string, instance any) *closeInstance {
	return &closeInstance{
		Name:     name,
		Instance: instance,
	}
}
