package v2

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Hook определяет тип функции-конфигуратора для применения хуков к структуре Hooks.
type Hook[E domain.Entity[ID], ID comparable] func(*Hooks[E, ID])

// Hooks содержит набор функций-колбэков (хуков) для кастомизации жизненного цикла
// CRUD-операций обобщенной сущности в слое доступа к данным (DAL).
type Hooks[E domain.Entity[ID], ID comparable] struct {
	// DispatchMapper отвечает за маппинг событий или диспетчеризацию данных.
	DispatchMapper DispatchMapperFunc[E, ID]
	// NewEntityFactory используется для создания новых экземпляров сущностей.
	NewEntityFactory NewEntityFactory[E, ID]

	// ValidateFind валидирует идентификатор перед поиском сущности.
	ValidateFind ValidateIDFunc[ID]
	// AfterFind вызывается после успешного извлечения сущности из репозитория.
	AfterFind AfterFindFunc[E, ID]

	// ListFilterYield обрабатывает фильтрацию при получении списка сущностей.
	ListFilterYield ListFilterYieldFunc[E, ID]

	// ValidateCreate валидирует сущность перед её созданием.
	ValidateCreate ValidateFunc[E, ID]
	// BeforeCreate выполняет подготовку данных или логику непосредственно перед созданием.
	BeforeCreate BeforeFunc[E, ID]
	// Creator выполняет кастомную операцию сохранения сущности в базу данных.
	Creator ModificatorFunc[E, ID]

	// ValidateChange валидирует измененную сущность перед обновлением.
	ValidateChange ValidateFunc[E, ID]
	// BeforeChange подготавливает данные или логику непосредственно перед обновлением.
	BeforeChange BeforeFunc[E, ID]
	// Changer выполняет кастомную операцию обновления сущности в базе данных.
	Changer ModificatorFunc[E, ID]

	// ValidateDelete валидирует идентификатор сущности перед её удалением.
	ValidateDelete ValidateIDFunc[ID]
}

// NewHooks создает и возвращает пустой экземпляр структуры Hooks.
func NewHooks[E domain.Entity[ID], ID comparable]() *Hooks[E, ID] {
	return &Hooks[E, ID]{}
}

// Validate проверяет наличие обязательных зависимостей в структуре Hooks.
// Возвращает ошибку, если не заданы DispatchMapper или NewEntityFactory.
func (ho *Hooks[E, ID]) Validate() error {
	if utils.IsNil(ho.DispatchMapper) {
		return errs.NewDalError("Hooks[E, ID].Validate", "DispatchMapper required", nil)
	}
	if utils.IsNil(ho.NewEntityFactory) {
		return errs.NewDalError("Hooks[E, ID].Validate", "NewEntityFactory required", nil)
	}

	return nil
}

// ===========================================================
// Fluent API (Функциональные опции для инициализации хуков)
// ===========================================================

// withHookDispatchMapper (неэкспортируемый) задает функцию маппинга/диспетчеризации.
func withHookDispatchMapper[E domain.Entity[ID], ID comparable](dispatcher DispatchMapperFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.DispatchMapper = dispatcher
	}
}

// WithHookNewEntityFactory задает фабрику для создания новых экземпляров сущностей.
func WithHookNewEntityFactory[E domain.Entity[ID], ID comparable](factory NewEntityFactory[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.NewEntityFactory = factory
	}
}

// WithValidateFind задает хук для валидации ID перед операцией поиска.
func WithValidateFind[E domain.Entity[ID], ID comparable](hook ValidateIDFunc[ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateFind = hook
	}
}

// WithHookAfterFind задает хук, выполняемый после успешного поиска сущности.
func WithHookAfterFind[E domain.Entity[ID], ID comparable](hook AfterFindFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.AfterFind = hook
	}
}

// WithHookListFilterYield задает фильтр для списков сущностей.
func WithHookListFilterYield[E domain.Entity[ID], ID comparable](hook ListFilterYieldFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ListFilterYield = hook
	}
}

// WithHookValidateCreate задает хук для валидации данных перед созданием.
func WithHookValidateCreate[E domain.Entity[ID], ID comparable](hook ValidateFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateCreate = hook
	}
}

// WithHookBeforeCreate задает хук, срабатывающий непосредственно перед сохранением новой сущности.
func WithHookBeforeCreate[E domain.Entity[ID], ID comparable](hook BeforeFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.BeforeCreate = hook
	}
}

// WithHookCreator задает кастомную функцию создания сущности в БД.
func WithHookCreator[E domain.Entity[ID], ID comparable](hook ModificatorFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.Creator = hook
	}
}

// WithHookValidateChange задает хук для валидации данных перед изменением.
func WithHookValidateChange[E domain.Entity[ID], ID comparable](hook ValidateFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateChange = hook
	}
}

// WithHookBeforeChange задает хук, срабатывающий непосредственно перед сохранением изменений сущности.
func WithHookBeforeChange[E domain.Entity[ID], ID comparable](hook BeforeFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.BeforeChange = hook
	}
}

// WithHookChanger задает кастомную функцию обновления сущности в БД.
func WithHookChanger[E domain.Entity[ID], ID comparable](hook ModificatorFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.Changer = hook
	}
}

// WithHookValidateDelete задает хук для валидации ID перед удалением.
func WithHookValidateDelete[E domain.Entity[ID], ID comparable](hook ValidateIDFunc[ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateDelete = hook
	}
}
