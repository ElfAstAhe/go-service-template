package usecase

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

// TailGetUseCase определяет обобщенный контракт оркестрации прикладного уровня для извлечения
// исторических цепочек или границ срезов логов, соответствующих определенным временным критериям.
type TailGetUseCase[ID comparable] interface {
	// GetTail возвращает срез идентификаторов записей, находящихся за границей указанной временной метки tail.
	GetTail(ctx context.Context, tail time.Time) ([]ID, error)
}

// TailGetInteractor оркестрирует несвязанные запросы поиска, управляя контрольными точками
// инвариантов состояния перед отправкой запросов в нижележащие обобщенные хранилища данных.
type TailGetInteractor[ID comparable] struct {
	tailRepo domain.TailRepository[ID] // Реализация нижележащего обобщенного слоя репозитория постоянного хранения
}

// Проверяем соответствие интерактора контракту UseCase на этапе компиляции для строковых ключей
var _ TailGetUseCase[string] = (*TailGetInteractor[string])(nil)

// NewTailGetUseCase — фабричный конструктор, инициализирующий зависимости адаптеров постоянного хранения данных.
func NewTailGetUseCase[ID comparable](tailRepo domain.TailRepository[ID]) *TailGetInteractor[ID] {
	return &TailGetInteractor[ID]{
		tailRepo: tailRepo,
	}
}

// GetTail оценивает инварианты временных границ перед делегированием диагностической команды поиска вниз по потоку репозитория.
func (tl *TailGetInteractor[ID]) GetTail(ctx context.Context, tail time.Time) ([]ID, error) {
	if tail.IsZero() {
		return nil, errs.NewInvalidArgumentError("tail", "must not be zero")
	}

	return tl.tailRepo.GetTail(ctx, tail)
}
