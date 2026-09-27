package repository

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// BaseRepositoryCallbacks инкапсулирует набор строго типизированных функций обратного вызова (Hooks/Lifecycles).
// Используется базовыми SQL-репозиториями для делегирования задач валидации,
// сканирования строк (Database Scanning), хуков жизненного цикла и фабричного создания сущностей доменного слоя.
type BaseRepositoryCallbacks[T domain.Entity[ID], ID comparable] struct {
	// EntityScanner отвечает за маппинг сырых строк базы данных (sql.Rows/sql.Row) в доменную структуру.
	EntityScanner EntityScannerFunc[T, ID]

	// NewEntityFactory — фабричный метод для генерации чистого экземпляра сущности в памяти.
	NewEntityFactory NewEntityFactory[T, ID]
	// AfterFind вызывается сразу после извлечения одиночной сущности из БД (например, для ленивой загрузки).
	AfterFind AfterFindFunc[T, ID]
	// AfterListYield выполняет пост-обработку среза (слайса) сущностей после массовой выборки.
	AfterListYield AfterListYieldFunc[T, ID]

	// ValidateCreate проверяет бизнес-инварианты сущности перед её вставкой в базу данных.
	ValidateCreate ValidateEntityFunc[T, ID]
	// Creator инкапсулирует кастомную SQL-логику вставки (INSERT), если дефолтный механизм не подходит.
	Creator CreatorFunc[T, ID]
	// BeforeCreate модифицирует сущность прямо перед записью (например, выставляет createdAt/UUID).
	BeforeCreate BeforeCreateFunc[T, ID]

	// ValidateChange проверяет инварианты сущности перед обновлением (UPDATE).
	ValidateChange ValidateEntityFunc[T, ID]
	// Changer инкапсулирует кастомную SQL-логику обновления записи.
	Changer ChangerFunc[T, ID]
	// BeforeChange модифицирует сущность перед апдейтом (например, обновляет updatedAt).
	BeforeChange BeforeChangeFunc[T, ID]
}

// newEmptyBaseRepositoryCallbacks генерирует пустой контейнер колбеков.
func newEmptyBaseRepositoryCallbacks[T domain.Entity[ID], ID comparable]() *BaseRepositoryCallbacks[T, ID] {
	return &BaseRepositoryCallbacks[T, ID]{}
}

// BaseRepositoryCallbacksBuilder реализует паттерн Строитель (Builder) для пошаговой и безопасной сборки хуков.
type BaseRepositoryCallbacksBuilder[T domain.Entity[ID], ID comparable] struct {
	instance *BaseRepositoryCallbacks[T, ID]
}

// NewBaseRepositoryCallbacksBuilder — фабричный конструктор строителя.
// ИСПРАВЛЕНО: Сразу аллоцирует instance, полностью защищая рантайм от паник nil pointer при вызове мутаторов.
func NewBaseRepositoryCallbacksBuilder[T domain.Entity[ID], ID comparable]() *BaseRepositoryCallbacksBuilder[T, ID] {
	return &BaseRepositoryCallbacksBuilder[T, ID]{
		instance: newEmptyBaseRepositoryCallbacks[T, ID](),
	}
}

// NewInstance принудительно сбрасывает состояние билдера, подкладывая чистую структуру.
func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) NewInstance() *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance = newEmptyBaseRepositoryCallbacks[T, ID]()
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithEntityScanner(scanner EntityScannerFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.EntityScanner = scanner
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithNewEntityFactory(factory NewEntityFactory[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.NewEntityFactory = factory
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithAfterFind(after AfterFindFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.AfterFind = after
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithAfterListYield(after AfterListYieldFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.AfterListYield = after
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithValidateCreate(validate ValidateEntityFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.ValidateCreate = validate
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithCreator(creator CreatorFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.Creator = creator
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithBeforeCreate(before BeforeCreateFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.BeforeCreate = before
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithValidateChange(validate ValidateEntityFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.ValidateChange = validate
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithChanger(changer ChangerFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.Changer = changer
	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithBeforeChange(before BeforeChangeFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.BeforeChange = before
	return bbr
}

// Build завершает сборку и возвращает полностью укомплектованный объект колбеков.
func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) Build() (*BaseRepositoryCallbacks[T, ID], error) {
	return bbr.instance, nil
}
