package v2

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// Scannable инкапсулирует абстракцию над структурами чтения строк БД (*sql.Row и *sql.Rows).
// Позволяет унифицировать парсинг результатов выборки, изолируя код от конкретных типов драйвера.
type Scannable interface {
	Scan(...any) error
	Err() error
}

// YieldAction определяет действие, которое необходимо совершить над сущностью в процессе итерации по списку.
type YieldAction int

const (
	// ListYieldActionAppend указывает на необходимость добавить текущую сущность в итоговую коллекцию.
	ListYieldActionAppend YieldAction = iota
	// ListYieldActionSkip указывает на необходимость пропустить текущую сущность без добавления в коллекцию.
	ListYieldActionSkip
)

type (
	// DispatchMapperFunc выполняет маппинг полей текущей строки базы данных в целевой доменный объект dest.
	DispatchMapperFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, mapperKey string, scannable Scannable, params ...any) (E, error)

	// NewEntityFactory — фабричный метод рантайма для генерации чистого (нулевого) экземпляра сущности T.
	NewEntityFactory[E domain.Entity[ID], ID comparable] func() E

	// EntityMapperFunc выполняет маппинг полей текущей строки базы данных в целевой доменный объект.
	EntityMapperFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, scannable Scannable, params ...any) (E, error)

	// AfterFindFunc — хук пост-обработки сущности сразу после её успешного извлечения из базы данных.
	AfterFindFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) (E, error)

	// ListFilterYieldFunc — хук-итератор, обрабатывающий каждую отдельную сущность в цикле вычитки слайса (List).
	ListFilterYieldFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) (E, YieldAction, error)

	// ValidateFunc выполняет валидацию перед фиксацией изменений.
	ValidateFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) error

	// BeforeFunc — хук пред-обработки сущности, выполняемый строго до совершения транзакции.
	BeforeFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, entity E, params ...any) (E, error)

	// ModificatorFunc инкапсулирует низкоуровневый SQL-вызов вставки/изменения записи, возвращая указатель на строку.
	ModificatorFunc[E domain.Entity[ID], ID comparable] func(ctx context.Context, querier db.Querier, entity E, params ...any) (Scannable, error)

	// ValidateIDFunc выполняет валидацию идентификатора
	ValidateIDFunc[ID comparable] func(ctx context.Context, id ID) error
)
