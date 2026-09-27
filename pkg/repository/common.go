package repository

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// LinkStrategy определяет архитектурную стратегию связывания доменных сущностей на уровне базы данных.
type LinkStrategy int

const (
	// LinkStrategyOneToMany — стратегия отношения «Один ко Многим» (внешний ключ на стороне дочерней таблицы).
	LinkStrategyOneToMany LinkStrategy = iota
	// LinkStrategyManyToMany — стратегия отношения «Многие ко Многим» (через промежуточную связующую таблицу-маппер).
	LinkStrategyManyToMany
)

// OwnedSaveFunc описывает функциональный контракт для пакетного сохранения/синхронизации зависимых (owned) сущностей.
type OwnedSaveFunc[T domain.Entity[ID], ID comparable, OwnerID comparable] func(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error)

// QueryBuilderFunc определяет сигнатуру функции-генератора динамических или статических SQL-запросов.
type QueryBuilderFunc func() string

// Scannable инкапсулирует абстракцию над структурами чтения строк БД (*sql.Row и *sql.Rows).
// Позволяет унифицировать парсинг результатов выборки, изолируя код от конкретных типов драйвера.
type Scannable interface {
	Scan(...any) error
}

// Блок строго типизированных сигнатур обратного вызова (Callbacks) для расширения базового SQL-движка.
type (
	// EntityScannerFunc выполняет маппинг полей текущей строки базы данных в целевой доменный объект dest.
	EntityScannerFunc[T domain.Entity[ID], ID comparable] func(scanner Scannable, sourceLabel string, dest T, params ...any) error

	// AfterFindFunc — хук пост-обработки сущности сразу после её успешного извлечения из базы данных.
	AfterFindFunc[T domain.Entity[ID], ID comparable] func(entity T, params ...any) (T, error)

	// AfterListYieldFunc — хук-итератор, обрабатывающий каждую отдельную сущность в цикле вычитки слайса (List).
	// Булев флаг позволяет прервать итерацию (yield) досрочно.
	AfterListYieldFunc[T domain.Entity[ID], ID comparable] func(entity T, params ...any) (T, bool, error)

	// NewEntityFactory — фабричный метод рантайма для генерации чистого (нулевого) экземпляра сущности T.
	NewEntityFactory[T domain.Entity[ID], ID comparable] func() T

	// ValidateEntityFunc выполняет валидацию инвариантов и ограничений сущности перед фиксацией изменений.
	ValidateEntityFunc[T domain.Entity[ID], ID comparable] func(entity T, params ...any) error

	// BeforeCreateFunc — хук пред-обработки сущности, выполняемый строго до совершения INSERT-транзакции.
	BeforeCreateFunc[T domain.Entity[ID], ID comparable] func(entity T, params ...any) error

	// BeforeChangeFunc — хук пред-обработки сущности, выполняемый строго до совершения UPDATE-транзакции.
	BeforeChangeFunc[T domain.Entity[ID], ID comparable] func(entity T, params ...any) error

	// CreatorFunc инкапсулирует низкоуровневый SQL-вызов вставки записи, возвращая указатель на строку сгенерированного ID.
	CreatorFunc[T domain.Entity[ID], ID comparable] func(ctx context.Context, querier db.Querier, entity T, params ...any) (*sql.Row, error)

	// ChangerFunc инкапсулирует низкоуровневый SQL-вызов модификации существующей записи.
	ChangerFunc[T domain.Entity[ID], ID comparable] func(ctx context.Context, querier db.Querier, entity T, params ...any) (*sql.Row, error)
)

// EntityInfo инкапсулирует строковые метаданные о соответствии доменной сущности и физической таблицы в СУБД.
type EntityInfo struct {
	Table  string // Имя таблицы в PostgreSQL (например, "audit_logs")
	Entity string // Имя доменной сущности в Go рантайме (например, "AuditLog")
}

// NewEntityInfo — конструктор метаданных сущности.
func NewEntityInfo(table, entity string) *EntityInfo {
	return &EntityInfo{
		Table:  table,
		Entity: entity,
	}
}
