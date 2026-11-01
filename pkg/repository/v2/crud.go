package v2

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const (
	// StatementFind идентификатор для операции поиска одиночной сущности (Find)
	StatementFind string = "statement-find"

	// StatementList идентификатор для операции выборки коллекций данных (List/All)
	StatementList string = "statement-list"

	// StatementCreate идентификатор для операции вставки записи (Create INSERT)
	StatementCreate string = "statement-create"

	// StatementChange идентификатор для операции модификации полей (Change UPDATE)
	StatementChange string = "statement-change"

	// StatementDelete идентификатор для операции жесткого или мягкого удаления (Delete)
	StatementDelete string = "statement-delete"
)

// CRUD реализует обобщенный репозиторий для выполнения базовых CRUD-операций над доменными сущностями.
// Объединяет в себе логику конфигурации, валидации, вызова хуков и исполнения SQL-запросов через Helper.
type CRUD[E domain.Entity[ID], ID comparable] struct {
	helper *Helper[E, ID]
	opts   *CRUDOptions[E, ID]
}

// Гарантия компиляции: проверка соответствия CRUD интерфейсу domain.CRUDRepository.
var _ domain.CRUDRepository[*dummyModel, string] = (*CRUD[*dummyModel, string])(nil)

// NewCRUD — фабричный метод создания экземпляра CRUD репозитория.
// Инициализирует опции, применяет хуки, строит SQL-выражения (statements) и настраивает внутренний Helper.
func NewCRUD[E domain.Entity[ID], ID comparable](options ...CRUDOption[E, ID]) (*CRUD[E, ID], error) {
	// Инициализация и применение переданных функциональных опций репозитория
	opts := NewCRUDOptions[E, ID]()
	for _, option := range options {
		option(opts)
	}
	if err := opts.Validate(); err != nil {
		return nil, errs.NewDalError("NewCRUD[E domain.Entity[ID]", "options validations failed", err)
	}

	// Создание основного экземпляра репозитория
	res := &CRUD[E, ID]{
		opts: opts,
	}

	// Регистрация хуков: добавляем внутренний mapperDispatcher к списку пользовательских хуков
	hookList := append(opts.Hooks, withHookDispatchMapper[E, ID](res.mapperDispatcher))
	hooks := NewHooks[E, ID]()
	for _, hook := range hookList {
		hook(hooks)
	}
	if err := hooks.Validate(); err != nil {
		return nil, errs.NewDalError("NewCRUD[E domain.Entity[ID]", "hooks validations failed", err)
	}

	// Сборка SQL-запросов (statements) под конкретный драйвер базы данных
	statements, err := opts.GetStatements(opts.DriverName)
	if err != nil {
		return nil, errs.NewDalError("NewCRUD[E domain.Entity[ID]", "statements build failed", err)
	}

	// Инициализация низкоуровневого хелпера для работы с БД и жизненным циклом сущности
	helper := NewHelper[E, ID](
		opts.Executor,
		opts.ErrorDecipher,
		hooks,
		statements,
		opts.TableName,
		opts.EntityName,
	)

	// Привязка настроенного хелпера к репозиторию
	res.helper = helper

	return res, nil
}

// Find выполняет поиск и извлечение одиночной доменной сущности по её уникаческому идентификатору.
// Перед отправкой запроса запускает внутренний хук валидации ValidateFind.
func (cr *CRUD[E, ID]) Find(ctx context.Context, id ID) (E, error) {
	if err := cr.validateFind(ctx, id); err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].Find", "find failed", err)
	}

	return cr.helper.Get(ctx, StatementFind, id)
}

// validateFind проверяет корректность ID перед поиском с помощью пользовательского хука.
func (cr *CRUD[E, ID]) validateFind(ctx context.Context, id ID) error {
	if !utils.IsNil(cr.helper.GetHooks().ValidateFind) {
		if err := cr.helper.GetHooks().ValidateFind(ctx, id); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateFind", "find validate hook failed", err)
		}
	}

	return nil
}

// List возвращает постраничную выборку (срез) доменных сущностей с учетом лимитов и смещений.
// Проверяет валидность параметров пагинации (limit > 0, offset >= 0).
func (cr *CRUD[E, ID]) List(ctx context.Context, limit, offset int) ([]E, error) {
	if err := cr.validateList(limit, offset); err != nil {
		return nil, errs.NewDalError("CRUD[E, ID].List", "list failed", err)
	}

	return cr.helper.List(ctx, StatementList, limit, offset)
}

// validateList выполняет базовую валидацию параметров лимита и смещения для списков.
func (cr *CRUD[E, ID]) validateList(limit, offset int) error {
	if limit <= 0 {
		return errs.NewDalError("CRUD[E, ID].validateList", "limit must be greater 0", nil)
	}
	if offset < 0 {
		return errs.NewDalError("CRUD[E, ID].validateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

// Create запускает валидацию бизнес-инвариантов и атомарно вставляет новую сущность в базу данных.
// Требует обязательного наличия хука Creator для сохранения сущности в БД.
func (cr *CRUD[E, ID]) Create(ctx context.Context, entity E) (E, error) {
	if err := cr.validateCreate(ctx, entity); err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].Create", "create entity failed", err)
	}

	return cr.helper.Create(ctx, StatementCreate, entity)
}

// validateCreate проверяет готовность сущности к созданию: наличие Creator, доменную валидацию и хук ValidateCreate.
func (cr *CRUD[E, ID]) validateCreate(ctx context.Context, entity E) error {
	// Проверка наличия хука модификации (без него запись невозможна)
	if utils.IsNil(cr.helper.GetHooks().Creator) {
		return errs.NewNotImplementedError(errs.NewDalError("CRUD[E, ID].validateCreate", "creator hook not applied", nil))
	}
	// Валидация на уровне самой доменной модели
	if err := entity.ValidateCreate(); err != nil {
		return errs.NewDalError("CRUD[E, ID].validateCreate", "entity validation failed", err)
	}
	// Вызов кастомного хука валидации процесса создания
	if !utils.IsNil(cr.helper.GetHooks().ValidateCreate) {
		if err := cr.helper.GetHooks().ValidateCreate(ctx, entity); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateCreate", "create validate hook failed", err)
		}
	}

	return nil
}

// Change выполняет обновление (UPDATE) полей существующей сущности.
// Требует наличия хука Changer для перезаписи данных в хранилище.
func (cr *CRUD[E, ID]) Change(ctx context.Context, entity E) (E, error) {
	if err := cr.validateChange(ctx, entity); err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].Change", "change entity failed", err)
	}

	return cr.helper.Change(ctx, StatementChange, entity)
}

// validateChange проверяет готовность сущности к изменению: наличие Changer, доменную валидацию и хук ValidateChange.
func (cr *CRUD[E, ID]) validateChange(ctx context.Context, entity E) error {
	// Проверка наличия хука модификации при изменении
	if utils.IsNil(cr.helper.GetHooks().Changer) {
		return errs.NewNotImplementedError(errs.NewDalError("CRUD[E, ID].validateChange", "changer hook not applied", nil))
	}
	// Валидация на уровне самой доменной модели перед обновлением
	if err := entity.ValidateChange(); err != nil {
		return errs.NewDalError("CRUD[E, ID].validateChange", "entity validation failed", err)
	}
	// Вызов кастомного хука валидации процесса изменения
	if !utils.IsNil(cr.helper.GetHooks().ValidateChange) {
		if err := cr.helper.GetHooks().ValidateChange(ctx, entity); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateChange", "change validate hook failed", err)
		}
	}

	return nil
}

// Delete удаляет запись из физической таблицы по её уникальному идентификатору ID.
// Запускает кастомную валидацию ValidateDelete перед фактическим стиранием.
func (cr *CRUD[E, ID]) Delete(ctx context.Context, id ID) error {
	if err := cr.validateDelete(ctx, id); err != nil {
		return errs.NewDalError("CRUD[E, ID].Delete", "delete failed", err)
	}

	return cr.helper.Delete(ctx, StatementDelete, id)
}

// validateDelete проверяет возможность удаления идентификатора с помощью хука.
func (cr *CRUD[E, ID]) validateDelete(ctx context.Context, id ID) error {
	if !utils.IsNil(cr.helper.GetHooks().ValidateDelete) {
		if err := cr.helper.GetHooks().ValidateDelete(ctx, id); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateDelete", "delete validate hook failed", err)
		}
	}

	return nil
}

// mapperDispatcher выполняет поиск маппера для указанного SQL-выражения и осуществляет сканирование строк БД в структуру доменной сущности.
func (cr *CRUD[E, ID]) mapperDispatcher(ctx context.Context, stmtName string, scannable Scannable, params ...any) (E, error) {
	mapper, err := cr.helper.GetStatements().GetMapper(stmtName)
	if err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].mapperDispatcher", fmt.Sprintf("mapper %s not found", stmtName), err)
	}

	return mapper(ctx, scannable, params...)
}

// GetOpts возвращает текущий набор опций конфигурации (CRUDOptions) репозитория.
func (cr *CRUD[E, ID]) GetOpts() *CRUDOptions[E, ID] {
	return cr.opts
}

// GetHelper возвращает внутренний экземпляр Helper, управляющий операциями репозитория.
func (cr *CRUD[E, ID]) GetHelper() *Helper[E, ID] {
	return cr.helper
}
