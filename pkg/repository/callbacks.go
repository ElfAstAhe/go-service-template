package repository

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// Callbacks инкапсулирует набор строго типизированных функций обратного вызова (Hooks/Lifecycles).
// Используется базовыми SQL-репозиториями для делегирования задач валидации,
// сканирования строк (Database Scanning), хуков жизненного цикла и фабричного создания сущностей доменного слоя.
type Callbacks[T domain.Entity[ID], ID comparable] struct {
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

// newEmptyCallbacks генерирует пустой контейнер колбеков для внутренней инициализации.
func newEmptyCallbacks[T domain.Entity[ID], ID comparable]() *Callbacks[T, ID] {
	return &Callbacks[T, ID]{}
}

// CallbacksBuilder реализует паттерн Строитель (Builder) для пошаговой и безопасной сборки хуков.
type CallbacksBuilder[T domain.Entity[ID], ID comparable] struct {
	instance *Callbacks[T, ID] // Удерживаемый в памяти собираемый экземпляр колбеков
}

// NewCallbacksBuilder — фабричный конструктор строителя.
// ИСПРАВЛЕНО: Сразу аллоцирует instance, полностью защищая рантайм от паник nil pointer при вызове мутаторов.
func NewCallbacksBuilder[T domain.Entity[ID], ID comparable]() *CallbacksBuilder[T, ID] {
	return &CallbacksBuilder[T, ID]{
		instance: newEmptyCallbacks[T, ID](),
	}
}

// NewInstance принудительно сбрасывает состояние билдера, подкладывая чистую структуру (мутатор повторного использования).
func (bbr *CallbacksBuilder[T, ID]) NewInstance() *CallbacksBuilder[T, ID] {
	bbr.instance = newEmptyCallbacks[T, ID]()
	return bbr
}

// WithEntityScanner инжектирует кастомную функцию сканирования строк СУБД.
func (bbr *CallbacksBuilder[T, ID]) WithEntityScanner(scanner EntityScannerFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.EntityScanner = scanner
	return bbr
}

// WithNewEntityFactory инжектирует доменную фабрику чистых сущностей.
func (bbr *CallbacksBuilder[T, ID]) WithNewEntityFactory(factory NewEntityFactory[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.NewEntityFactory = factory
	return bbr
}

// WithAfterFind инжектирует хук пост-обработки сущности после одиночного поиска.
func (bbr *CallbacksBuilder[T, ID]) WithAfterFind(after AfterFindFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.AfterFind = after
	return bbr
}

// WithAfterListYield инжектирует хук пакетной пост-обработки сущностей после выборок.
func (bbr *CallbacksBuilder[T, ID]) WithAfterListYield(after AfterListYieldFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.AfterListYield = after
	return bbr
}

// WithValidateCreate инжектирует функцию валидации инвариантов перед операцией INSERT.
func (bbr *CallbacksBuilder[T, ID]) WithValidateCreate(validate ValidateEntityFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.ValidateCreate = validate
	return bbr
}

// WithCreator инжектирует кастомный SQL-исполнитель операции INSERT.
func (bbr *CallbacksBuilder[T, ID]) WithCreator(creator CreatorFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.Creator = creator
	return bbr
}

// WithBeforeCreate инжектирует хук модификации полей сущности перед операцией INSERT.
func (bbr *CallbacksBuilder[T, ID]) WithBeforeCreate(before BeforeCreateFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.BeforeCreate = before
	return bbr
}

// WithValidateChange инжектирует функцию валидации инвариантов перед операцией UPDATE.
func (bbr *CallbacksBuilder[T, ID]) WithValidateChange(validate ValidateEntityFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.ValidateChange = validate
	return bbr
}

// WithChanger инжектирует кастомный SQL-исполнитель операции UPDATE.
func (bbr *CallbacksBuilder[T, ID]) WithChanger(changer ChangerFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.Changer = changer
	return bbr
}

// WithBeforeChange инжектирует хук модификации полей сущности перед операцией UPDATE.
func (bbr *CallbacksBuilder[T, ID]) WithBeforeChange(before BeforeChangeFunc[T, ID]) *CallbacksBuilder[T, ID] {
	bbr.instance.BeforeChange = before
	return bbr
}

// Build завершает пошаговую сборку и возвращает полностью укомплектованный объект колбеков.
func (bbr *CallbacksBuilder[T, ID]) Build() (*Callbacks[T, ID], error) {
	return bbr.instance, nil
}
