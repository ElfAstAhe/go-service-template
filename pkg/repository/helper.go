package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Helper — инфраструктурный движок (низкоуровневый оркестратор) репозиторного слоя.
//
// Инкапсулирует рутину выполнения SQL-запросов, автоматическое извлечение Querier
// из контекста транзакций (ACID) и последовательный запуск хуков жизненного цикла сущностей.
type Helper[T domain.Entity[ID], ID comparable] struct {
	exec        db.Executor                     // Менеджер пула соединений с БД
	errDecipher db.ErrorDecipher                // Переводчик системных ошибок СУБД в ошибки фреймворка
	info        *EntityInfo                     // Метаданные о сущности и таблице
	nilInstance T                               // Пустой литерал типа T для безопасных возвратов при сбоях
	callbacks   *BaseRepositoryCallbacks[T, ID] // Контейнер хуков и сканеров жизненного цикла
}

// newHelper — фабричный конструктор низкоуровневого хелпера репозитория.
func newHelper[T domain.Entity[ID], ID comparable](exec db.Executor, errDecipher db.ErrorDecipher, callbacks *BaseRepositoryCallbacks[T, ID], info *EntityInfo) *Helper[T, ID] {
	return &Helper[T, ID]{
		exec:        exec,
		errDecipher: errDecipher,
		info:        info,
		callbacks:   callbacks,
	}
}

// ====================================================================
// Публичные геттеры зависимостей рантайма
// ====================================================================

func (h *Helper[T, ID]) GetExecutor() db.Executor                      { return h.exec }
func (h *Helper[T, ID]) GetErrDecipher() db.ErrorDecipher              { return h.errDecipher }
func (h *Helper[T, ID]) GetInfo() *EntityInfo                          { return h.info }
func (h *Helper[T, ID]) GetNilInstance() T                             { return h.nilInstance }
func (h *Helper[T, ID]) GetCallbacks() *BaseRepositoryCallbacks[T, ID] { return h.callbacks }

// Get выполняет низкоуровневое чтение одной строки, сканирование полей и запуск пост-хука AfterFind.
func (h *Helper[T, ID]) Get(ctx context.Context, sourceLabel string, sqlReq string, params ...any) (T, error) {
	// Автоматически извлекаем активный Querier (текущую транзакцию Tx или голый пул DB)
	querier := h.exec.GetQuerier(ctx)

	row := querier.QueryRowContext(ctx, sqlReq, params...)
	res := h.callbacks.NewEntityFactory()

	// Сканируем запись через зарегистрированный колбек маппинга
	err := h.callbacks.EntityScanner(row, sourceLabel, res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return h.nilInstance, errs.NewDalNotFoundError(h.info.Entity, params, err)
		}
		return h.nilInstance, errs.NewDalError("Helper.Get", "get row failed", err)
	}

	// Если определен хук пост-обработки — обогащаем сущность перед отдачей наверх
	if h.callbacks.AfterFind != nil {
		return h.callbacks.AfterFind(res, params...)
	}

	return res, nil
}

// List выполняет потоковую вычитку набора строк СУБД с защитой планировщика от отмены контекста.
func (h *Helper[T, ID]) List(ctx context.Context, sourceLabel string, sqlReq string, params ...any) ([]T, error) {
	querier := h.GetExecutor().GetQuerier(ctx)

	rows, err := querier.QueryContext(ctx, sqlReq, params...)
	if err != nil {
		return nil, errs.NewDalError("Helper.List", "query execution failed", err)
	}
	defer rows.Close()

	res := make([]T, 0)
	for rows.Next() {
		// Предохранитель: останавливаем итерации, если вышестоящий контекст (например, HTTP) был отменен [1]
		if err = ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return res, nil
			}
			return nil, errs.NewDalError("Helper.List", "got context error during iteration", err)
		}

		isAddEntity := true
		entity := h.GetCallbacks().NewEntityFactory()

		// Маппим текущую строку выборки в объект
		err = h.GetCallbacks().EntityScanner(rows, sourceLabel, entity, params...)
		if err != nil {
			return nil, errs.NewDalError("Helper.List", "scan rows failed", err)
		}

		// Вызываем хук фильтрации/модификации строки «на лету» (Yielder)
		if h.GetCallbacks().AfterListYield != nil {
			entity, isAddEntity, err = h.GetCallbacks().AfterListYield(entity, params...)
			if err != nil {
				return nil, errs.NewDalError("Helper.List", "post scan row yield failed", err)
			}
		}

		// Если yielder стер сущность или сбросил флаг добавления — пропускаем шаг
		if any(entity) == nil || !isAddEntity {
			continue
		}

		res = append(res, entity)
	}

	// Проверяем, не прервался ли цикл rows.Next() из-за системной ошибки драйвера
	if rows.Err() != nil {
		return nil, errs.NewDalError("Helper.List", "after scan rows failed", rows.Err())
	}

	return res, nil
}

// Create координирует вставку записи: вызывает пред-хуки, выполняет инжектированный Creator и проверяет уникальность ID.
func (h *Helper[T, ID]) Create(ctx context.Context, sourceLabel string, entity T, params ...any) (T, error) {
	if h.GetCallbacks().BeforeCreate != nil {
		if err := h.GetCallbacks().BeforeCreate(entity, params...); err != nil {
			return h.GetNilInstance(), errs.NewDalError("Helper.Create", "before create hook failed", err)
		}
	}

	querier := h.GetExecutor().GetQuerier(ctx)

	// Запускаем пользовательскую SQL-функцию вставки
	row, err := h.GetCallbacks().Creator(ctx, querier, entity, params...)
	if err != nil {
		return h.GetNilInstance(), errs.NewDalError("Helper.Create", "create entity execution failed", err)
	}

	res := h.GetCallbacks().NewEntityFactory()
	err = h.GetCallbacks().EntityScanner(row, sourceLabel, res, params...)
	if err != nil {
		// Дешифруем ошибку дублирования уникального индекса (Unique Violation)
		if h.errDecipher.IsUniqueViolation(err) {
			return h.GetNilInstance(), errs.NewDalAlreadyExistsError(h.GetInfo().Entity, entity.GetID(), err)
		}
		return h.GetNilInstance(), errs.NewDalError("Helper.Create", "scan after create entity failed", err)
	}

	if h.GetCallbacks().AfterFind != nil {
		return h.GetCallbacks().AfterFind(res, params...)
	}

	return res, nil
}

// Change координирует обновление параметров записи: запускает пред-хуки и выполняет инжектированный Changer.
func (h *Helper[T, ID]) Change(ctx context.Context, sourceLabel string, entity T, params ...any) (T, error) {
	if h.GetCallbacks().BeforeChange != nil {
		if err := h.GetCallbacks().BeforeChange(entity, params...); err != nil {
			return h.GetNilInstance(), errs.NewDalError("Helper.Change", "before change hook failed", err)
		}
	}

	querier := h.GetExecutor().GetQuerier(ctx)

	// Запускаем пользовательскую SQL-функцию апдейта
	row, err := h.GetCallbacks().Changer(ctx, querier, entity, params...)
	if err != nil {
		return h.GetNilInstance(), errs.NewDalError("Helper.Change", "change entity execution failed", err)
	}

	res := h.GetCallbacks().NewEntityFactory()
	err = h.GetCallbacks().EntityScanner(row, sourceLabel, res, params...)
	if err != nil {
		if h.errDecipher.IsUniqueViolation(err) {
			return h.GetNilInstance(), errs.NewDalAlreadyExistsError(h.GetInfo().Entity, entity.GetID(), err)
		}
		return h.GetNilInstance(), errs.NewDalError("Helper.Change", "scan after change entity failed", err)
	}

	if h.GetCallbacks().AfterFind != nil {
		return h.GetCallbacks().AfterFind(res, params...)
	}

	return res, nil
}

// Delete удаляет запись с обязательной жесткой проверкойRowsAffected (вернет ошибку, если запись не существовала).
func (h *Helper[T, ID]) Delete(ctx context.Context, sqlReq string, params ...any) error {
	querier := h.GetExecutor().GetQuerier(ctx)
	res, err := querier.ExecContext(ctx, sqlReq, params...)
	if err != nil {
		return errs.NewDalError("Helper.Delete", "exec context failed", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errs.NewDalError("Helper.Delete", "rows affected read failed", err)
	}
	if !(rowsAffected > 0) {
		return errs.NewDalNotFoundError(h.GetInfo().Entity, params, nil)
	}

	return nil
}

// DeleteNoCheck удаляет запись без верификации её предварительного существования (слепое удаление).
func (h *Helper[T, ID]) DeleteNoCheck(ctx context.Context, sqlReq string, params ...any) error {
	querier := h.GetExecutor().GetQuerier(ctx)
	_, err := querier.ExecContext(ctx, sqlReq, params...)
	if err != nil {
		return errs.NewDalError("Helper.DeleteNoCheck", "exec context failed", err)
	}

	return nil
}
