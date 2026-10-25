package v2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Helper — инфраструктурный движок (низкоуровневый оркестратор) репозиторного слоя.
//
// Инкапсулирует рутину выполнения SQL-запросов, автоматическое извлечение Querier
// из контекста транзакций (ACID) и последовательный запуск хуков жизненного цикла сущностей.
type Helper[E domain.Entity[ID], ID comparable] struct {
	executor    db.Executor        // Менеджер пула соединений с БД
	errDecipher db.ErrorDecipher   // Переводчик системных ошибок СУБД в ошибки фреймворка
	hooks       *Hooks[E, ID]      // Контейнер хуков и сканеров жизненного цикла
	statements  *Statements[E, ID] // Контейнер sql + mappers
	nilInstance E                  // Пустой (nil) литерал типа T для безопасных возвратов при сбоях
	tableName   string             // Наименование таблицы
	entityName  string             // Наименование entity
}

func NewHelper[E domain.Entity[ID], ID comparable](
	executor db.Executor,
	errDecipher db.ErrorDecipher,
	hooks *Hooks[E, ID],
	statements *Statements[E, ID],
	tableName string,
	entityName string,
) *Helper[E, ID] {
	return &Helper[E, ID]{
		executor:    executor,
		errDecipher: errDecipher,
		hooks:       hooks,
		statements:  statements,
		tableName:   tableName,
		entityName:  entityName,
	}
}

// GetExecutor возвращает инжектированную реализацию db.Executor для управления подключениями к СУБД.
func (h *Helper[E, ID]) GetExecutor() db.Executor { return h.executor }

// GetErrDecipher возвращает переводчик системных ошибок СУБД в структурированные исключения фреймворка.
func (h *Helper[E, ID]) GetErrDecipher() db.ErrorDecipher { return h.errDecipher }

// GetNilInstance возвращает дефолтный пустой литерал типа T для безопасных пустых возвратов при ошибках.
func (h *Helper[E, ID]) GetNilInstance() E { return h.nilInstance }

// GetHooks возвращает контейнер зарегистрированных хуков жизненного цикла Hooks.
func (h *Helper[E, ID]) GetHooks() *Hooks[E, ID] { return h.hooks }

// GetStatements возвращает контейнер sql + mapper
func (h *Helper[E, ID]) GetStatements() *Statements[E, ID] { return h.statements }

// Get выполняет точечное чтение одной строки (SELECT BY ID) из СУБД.
// Автоматически извлекает активный Querier из контекста транзакции, создаёт пустую заготовку сущности,
// осуществляет сканирование записи и, если настроен, запускает пост-хук AfterFind.
// Возвращает DalNotFoundError при отсутствии записи (sql.ErrNoRows).
func (h *Helper[E, ID]) Get(ctx context.Context, stmtName string, params ...any) (E, error) {
	querier := h.executor.GetQuerier(ctx)

	sqlStmt, err := h.statements.GetSQL(stmtName)
	if err != nil {
		return h.nilInstance, errs.NewDalError("Helper[E, ID].Get", fmt.Sprintf("sql statement %s not found", stmtName), err)
	}

	row := querier.QueryRowContext(ctx, sqlStmt, params...)

	res, err := h.hooks.DispatchMapper(ctx, stmtName, row, params...)

	//    err := h.callbacks.EntityScanner(row, sourceLabel, res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return h.nilInstance, errs.NewDalNotFoundError(h.entityName, params, err)
		}

		return h.nilInstance, errs.NewDalError("Helper[E, ID].Get", "get row failed", err)
	}

	if !utils.IsNil(h.hooks.AfterFind) {
		return h.hooks.AfterFind(ctx, res, params...)
	}

	return res, nil
}

// List выполняет потоковую вычитку набора строк СУБД (SELECT LIST).
// Защищает CPU планировщика от отмены контекста (проверяет ctx.Err на каждой итерации rows.Next).
// Вызывает EntityScanner для каждой строки и пропускает результаты через AfterListYield для фильтрации/модификации.
func (h *Helper[E, ID]) List(ctx context.Context, stmtName string, params ...any) ([]E, error) {
	querier := h.GetExecutor().GetQuerier(ctx)

	sqlStmt, err := h.statements.GetSQL(stmtName)
	if err != nil {
		return nil, errs.NewDalError("Helper[E, ID].List", fmt.Sprintf("sql statement %s not found", stmtName), err)
	}

	rows, err := querier.QueryContext(ctx, sqlStmt, params...)
	if err != nil {
		return nil, errs.NewDalError("Helper[E, ID].List", "query execution failed", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	res := make([]E, 0)
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return res, nil
			}
			return nil, errs.NewDalError("Helper[E, ID].List", "got context error during iteration", err)
		}

		appendEntity := ListYieldActionAppend
		//        entity := h.GetCallbacks().NewEntityFactory()

		//        err = h.GetCallbacks().EntityScanner(rows, sourceLabel, entity, params...)
		entity, err := h.hooks.DispatchMapper(ctx, stmtName, rows, params...)
		if err != nil {
			return nil, errs.NewDalError("Helper[E, ID].List", "scan rows failed", err)
		}

		if !utils.IsNil(h.hooks.ListFilterYield) {
			entity, appendEntity, err = h.hooks.ListFilterYield(ctx, entity, params...)
			if err != nil {
				return nil, errs.NewDalError("Helper[E, ID].List", "post scan row yield failed", err)
			}
		}

		if any(entity) == nil || appendEntity == ListYieldActionSkip {
			continue
		}

		res = append(res, entity)
	}

	if rows.Err() != nil {
		return nil, errs.NewDalError("Helper[E, ID].List", "after scan rows failed", rows.Err())
	}

	return res, nil
}

// Create координирует транзакционную вставку (INSERT) новой записи в базу данных.
// Последовательно выполняет пред-хуки BeforeCreate, запускает инжектированный SQL-замыкание Creator,
// парсит возвращенные автогенерируемые ID и дешифрует вендоро-специфичные ошибки на предмет коллизий уникальности.
func (h *Helper[E, ID]) Create(ctx context.Context, stmtName string, entity E, params ...any) (E, error) {
	var ent = entity
	var err error
	if !utils.IsNil(h.hooks.BeforeCreate) {
		if ent, err = h.hooks.BeforeCreate(ctx, entity, params...); err != nil {
			return h.GetNilInstance(), errs.NewDalError("Helper[T, ID].Create", "before create hook failed", err)
		}
	}

	querier := h.GetExecutor().GetQuerier(ctx)

	row, err := h.hooks.Creator(ctx, querier, ent, params...)
	if err != nil {
		return h.GetNilInstance(), errs.NewDalError("Helper[T, ID].Create", "create entity execution failed", err)
	}

	res, err := h.hooks.DispatchMapper(ctx, stmtName, row, params...)
	if err != nil {
		if h.errDecipher.IsUniqueViolation(err) {
			return h.GetNilInstance(), errs.NewDalAlreadyExistsError(h.entityName, entity.GetID(), err)
		}

		return h.GetNilInstance(), errs.NewDalError("Helper[T, ID].Create", "scan after create entity failed", err)
	}

	if !utils.IsNil(h.hooks.AfterFind) {
		return h.hooks.AfterFind(ctx, res, params...)
	}

	return res, nil
}

// Change координирует транзакционное обновление (UPDATE) полей существующей доменной записи.
// Последовательно запускает пред-хуки BeforeChange, выполняет прикладной SQL-код Changer
// и осуществляет повторное сканирование обновленных полей с обработкой коллизий уникальных индексов.
func (h *Helper[E, ID]) Change(ctx context.Context, stmtName string, entity E, params ...any) (E, error) {
	var ent = entity
	var err error
	if !utils.IsNil(h.hooks.BeforeChange) {
		if ent, err = h.hooks.BeforeChange(ctx, entity, params...); err != nil {
			return h.GetNilInstance(), errs.NewDalError("Helper[E, ID].Change", "before change hook failed", err)
		}
	}

	querier := h.GetExecutor().GetQuerier(ctx)

	row, err := h.hooks.Changer(ctx, querier, ent, params...)
	if err != nil {
		return h.GetNilInstance(), errs.NewDalError("Helper[E, ID].Change", "change entity execution failed", err)
	}

	res, err := h.hooks.DispatchMapper(ctx, stmtName, row, params...)
	if err != nil {
		if h.errDecipher.IsUniqueViolation(err) {
			return h.GetNilInstance(), errs.NewDalAlreadyExistsError(h.entityName, entity.GetID(), err)
		}

		return h.GetNilInstance(), errs.NewDalError("Helper[E, ID].Change", "scan after change entity failed", err)
	}

	if !utils.IsNil(h.hooks.AfterFind) {
		return h.hooks.AfterFind(ctx, res, params...)
	}

	return res, nil
}

// Delete выполняет удаление (DELETE) записи по переданному SQL-шаблону с обязательным аудитом измененных строк.
// Возвращает DalNotFoundError, если ни одна строка в таблице не была затронута (запись изначально отсутствовала).
//
//goland:noinspection GoLinter
func (h *Helper[E, ID]) Delete(ctx context.Context, stmtName string, params ...any) error {
	querier := h.GetExecutor().GetQuerier(ctx)

	sqlStmt, err := h.statements.GetSQL(stmtName)
	if err != nil {
		return errs.NewDalError("Helper[E, ID].Delete", fmt.Sprintf("sql statement %s not found", stmtName), err)
	}

	res, err := querier.ExecContext(ctx, sqlStmt, params...)
	if err != nil {
		return errs.NewDalError("Helper[E, ID].Delete", "exec context failed", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errs.NewDalError("Helper[E, ID].Delete", "rows affected read failed", err)
	}
	if rowsAffected <= 0 {
		return errs.NewDalNotFoundError(h.entityName, params, nil)
	}

	return nil
}

// DeleteNoCheck осуществляет быстрое «слепое» удаление (DELETE).
func (h *Helper[E, ID]) DeleteNoCheck(ctx context.Context, stmtName string, params ...any) error {
	querier := h.GetExecutor().GetQuerier(ctx)

	sqlStmt, err := h.statements.GetSQL(stmtName)
	if err != nil {
		return errs.NewDalError("Helper[E, ID].DeleteNoCheck", fmt.Sprintf("sql statement %s not found", stmtName), err)
	}

	_, err = querier.ExecContext(ctx, sqlStmt, params...)
	if err != nil {
		return errs.NewDalError("Helper[E, ID].DeleteNoCheck", "exec context failed", err)
	}

	return nil
}
