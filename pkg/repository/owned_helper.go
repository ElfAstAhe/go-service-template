package repository

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// OwnedHelper — специализированное расширение базового инфраструктурного движка Helper
// для работы с дочерними (owned) сущностями и пакетными выборками в разрезе владельцев.
//
// Использует композицию (встраивание структуры *Helper), наследуя всю готовую логику
// управления пулом соединений, дешифрации ошибок и атомарных CRUD-операций.
type OwnedHelper[T domain.Entity[ID], ID comparable, OwnerID comparable] struct {
	*Helper[T, ID] // Встраиваем базовый хелпер для переиспользования кода
}

// newOwnedHelper — фабричный конструктор расширенного хелпера для Owned-репозиториев.
func newOwnedHelper[T domain.Entity[ID], ID comparable, OwnerID comparable](exec db.Executor, errDecipher db.ErrorDecipher, callbacks *BaseRepositoryCallbacks[T, ID], info *EntityInfo) *OwnedHelper[T, ID, OwnerID] {
	return &OwnedHelper[T, ID, OwnerID]{
		Helper: newHelper[T, ID](exec, errDecipher, callbacks, info),
	}
}

// ListByOwners выполняет высокопроизводительную пакетную выборку (Bulk Fetching) данных
// для группы владельцев за один SQL-запрос, агрегируя результат в типизированную мапу слайсов.
// Позволяет полностью победить проблему N+1 запросов на уровне инфраструктуры фреймворка.
func (oh *OwnedHelper[T, ID, OwnerID]) ListByOwners(ctx context.Context, sourceLabel string, sqlReq string, params ...any) (map[OwnerID][]T, error) {
	querier := oh.GetExecutor().GetQuerier(ctx)

	rows, err := querier.QueryContext(ctx, sqlReq, params...)
	if err != nil {
		return nil, errs.NewDalError("OwnedHelper.ListByOwners", "query rows failed", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	res := make(map[OwnerID][]T)
	for rows.Next() {
		// Предохранитель планировщика: прерываем итерации, если контекст (HTTP/gRPC/Timeout) отменился
		if err = ctx.Err(); err != nil {
			// Если контекст отменен — возвращаем то, что успели накопить, без генерации ошибки
			if errors.Is(err, context.Canceled) {
				return res, nil
			}
			return nil, errs.NewDalError("OwnedHelper.ListByOwners", "got context error during batch iteration", err)
		}

		isAddEntity := true
		var ownerID OwnerID
		entity := oh.GetCallbacks().NewEntityFactory()

		// Сканируем строку, передавая указатель &ownerID для динамического определения владельца записи
		err = oh.GetCallbacks().EntityScanner(rows, sourceLabel, entity, &ownerID)
		if err != nil {
			return nil, errs.NewDalError("OwnedHelper.ListByOwners", "scan row failed", err)
		}

		// Вызываем хук фильтрации/модификации строки «на лету» (Yielder)
		if oh.GetCallbacks().AfterListYield != nil {
			entity, isAddEntity, err = oh.GetCallbacks().AfterListYield(entity, ownerID)
			if err != nil {
				return nil, errs.NewDalError("OwnedHelper.ListByOwners", "post scan row yield failed", err)
			}
		}

		// Если yielder отфильтровал (стер) сущность или сбросил флаг — пропускаем шаг
		if any(entity) == nil || !isAddEntity {
			continue
		}

		// Ленивая инициализация вложенного слайса для нового OwnerID (защита от nil pointer)
		if _, ok := res[ownerID]; !ok {
			res[ownerID] = make([]T, 0)
		}
		res[ownerID] = append(res[ownerID], entity)
	}

	// Финальная проверка на системный сбой сетевого драйвера в процессе итерации rows.Next()
	if rows.Err() != nil {
		return nil, errs.NewDalError("OwnedHelper.ListByOwners", "after scan rows failed", rows.Err())
	}

	return res, nil
}
