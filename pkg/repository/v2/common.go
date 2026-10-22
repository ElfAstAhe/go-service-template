package v2

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

// Scannable инкапсулирует абстракцию над структурами чтения строк БД (*sql.Row и *sql.Rows).
// Позволяет унифицировать парсинг результатов выборки, изолируя код от конкретных типов драйвера.
type Scannable interface {
	Scan(...any) error
}

type (
	// DispatchMapperFunc выполняет маппинг полей текущей строки базы данных в целевой доменный объект dest.
	DispatchMapperFunc[E domain.Entity[ID], ID comparable] func(context.Context, string, Scannable) (E, []any, error)

	// NewEntityFactory — фабричный метод рантайма для генерации чистого (нулевого) экземпляра сущности T.
	NewEntityFactory[E domain.Entity[ID], ID comparable] func() E

	// EntityMapperFunc выполняет маппинг полей текущей строки базы данных в целевой доменный объект.
	EntityMapperFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, scannable repository.Scannable, params ...any) (E, []any, error)

	// AfterFindFunc — хук пост-обработки сущности сразу после её успешного извлечения из базы данных.
	AfterFindFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) (E, error)

	// AfterListYieldFunc — хук-итератор, обрабатывающий каждую отдельную сущность в цикле вычитки слайса (List).
	// Булев флаг позволяет прервать итерацию (yield) досрочно.
	AfterListYieldFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) (E, bool, error)

	// ValidateEntityFunc выполняет валидацию инвариантов и ограничений сущности перед фиксацией изменений.
	ValidateEntityFunc[E domain.Entity[ID], ID comparable] func(entity E, params ...any) error

	// BeforeCreateFunc — хук пред-обработки сущности, выполняемый строго до совершения INSERT-транзакции.
	BeforeCreateFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) error

	// BeforeChangeFunc — хук пред-обработки сущности, выполняемый строго до совершения UPDATE-транзакции.
	BeforeChangeFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) error

	// CreatorFunc инкапсулирует низкоуровневый SQL-вызов вставки записи, возвращая указатель на строку сгенерированного ID.
	CreatorFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, querier db.Querier, entity E, params ...any) (*sql.Row, error)

	// ChangerFunc инкапсулирует низкоуровневый SQL-вызов модификации существующей записи.
	ChangerFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, querier db.Querier, entity E, params ...any) (*sql.Row, error)
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
