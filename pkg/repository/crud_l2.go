package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseCRUDL2Repository реализует паттерн проектирования Декоратор поверх стандартного CRUD-репозитория,
// обеспечивая сквозное кэширование второго уровня (Cache-Aside / Read-Through / Write-Through Caching Pattern).
//
// Инкапсулирует в себе логику проверки наличия данных во внешнем кэш-хранилище (Cache Hit) перед непосредственным
// обращением к СУБД (Cache Miss), а также берет на себя сквозную инвалидацию и обновление кэш-записей при мутациях.
type BaseCRUDL2Repository[E domain.Entity[ID], ID comparable] struct {
	next       domain.CRUDRepository[E, ID] // Ссылка на декорируемый репозиторий первого уровня (DAL СУБД)
	entityInfo *EntityInfo                  // Метаданные о доменной сущности и целевой таблице базы данных
	crudCache  cache.Cache[ID, E]           // Потокобезопасный интерфейс распределенного или локального кэш-драйвера
	nilEntity  E                            // Пустой литерал типа E для безопасных дефолтных возвратов при сбоях
	defaultTTL time.Duration                // Стандартное время жизни кэш-записей (Time-To-Live) в хранилище
	log        logger.Logger                // Изолированный структурированный логгер репозитория
}

// NewBaseCRUDL2Repository — фабричный конструктор репозитория сквозного кэширования второго уровня.
func NewBaseCRUDL2Repository[E domain.Entity[ID], ID comparable](
	next domain.CRUDRepository[E, ID],
	entityInfo *EntityInfo,
	crudCache cache.Cache[ID, E],
	defaultTTL time.Duration,
	log logger.Logger,
) *BaseCRUDL2Repository[E, ID] {
	return &BaseCRUDL2Repository[E, ID]{
		next:       next,
		entityInfo: entityInfo,
		crudCache:  crudCache,
		defaultTTL: defaultTTL,
		log:        log.GetLogger("BaseCRUDL2Repository"),
	}
}

// Find осуществляет точечный поиск сущности по ее уникальному идентификатору с использованием кэш-слоя.
// Сначала пытается извлечь объект из кэша. В случае промаха (Cache Miss) обращается к базе данных первого уровня,
// после чего детерминированно сохраняет результат выборки в кэш с заданным временем жизни defaultTTL.
func (bcl *BaseCRUDL2Repository[E, ID]) Find(ctx context.Context, id ID) (E, error) {
	// Шаг 1: Попытка быстрого извлечения объекта из оперативной памяти/Redis (Cache-Aside)
	res, ok, err := bcl.crudCache.Get(id)
	if err != nil {
		return bcl.nilEntity, errs.NewDalCacheError("BaseCRUDL2Repository.Find", fmt.Sprintf("get from cache entity id [%v]", id), err)
	}
	if ok {
		// Оборонительная проверка на логическое отсутствие записи (Negative Caching protection)
		if utils.IsNil(res) {
			return res, errs.NewDalNotFoundError(bcl.GetInfo().Entity, "not found", nil)
		}

		return res, nil // Успешное попадание в кэш (Cache Hit)
	}

	// Шаг 2: Промах кэша (Cache Miss) — выполняем оригинальную операцию чтения из реляционной СУБД
	res, err = bcl.next.Find(ctx, id)
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); !ok {
			return res, err
		}
	}

	// Шаг 3: Асинхронно-безопасное наполнение кэш-слоя актуальными данными
	cacheErr := bcl.crudCache.Set(id, res, bcl.defaultTTL)
	if cacheErr != nil {
		// Отказоустойчивость: сбой кэша не должен ломать транзакцию чтения, только логируем варнинг
		bcl.log.Errorf(fmt.Sprintf("set into cache entity id [%v]", id), cacheErr)
	}

	return res, nil
}

// List выполняет потоковое пакетное извлечение набора записей.
// В текущей реализации метод прозрачно делегирует вызов репозиторию первого уровня без кэширования срезов.
func (bcl *BaseCRUDL2Repository[E, ID]) List(ctx context.Context, limit, offset int) ([]E, error) {
	return bcl.next.List(ctx, limit, offset)
}

// Create координирует транзакционную вставку (INSERT) новой записи.
// После успешной фиксации строки в СУБД осуществляет превентивное наполнение кэша (Write-Through паттерн).
func (bcl *BaseCRUDL2Repository[E, ID]) Create(ctx context.Context, entity E) (E, error) {
	res, err := bcl.next.Create(ctx, entity)
	if err != nil {
		return bcl.nilEntity, err
	}

	// Синхронизируем состояние распределенной памяти с новыми данными СУБД
	cacheErr := bcl.crudCache.Set(res.GetID(), res, bcl.defaultTTL)
	if cacheErr != nil {
		bcl.log.Errorf(fmt.Sprintf("set into cache entity id [%v]", res.GetID()), cacheErr)
	}

	return res, err
}

// Change координирует обновление параметров записи (UPDATE) в СУБД.
// Автоматически обновляет или перезаписывает соответствующий ключ в кэше актуальным состоянием сущности.
func (bcl *BaseCRUDL2Repository[E, ID]) Change(ctx context.Context, entity E) (E, error) {
	res, err := bcl.next.Change(ctx, entity)
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); !ok {
			return res, err
		}
	}

	// Предотвращаем появление "протухших" данных (Stale Data), перезаписывая ключ в памяти
	cacheErr := bcl.crudCache.Set(res.GetID(), res, bcl.defaultTTL)
	if cacheErr != nil {
		bcl.log.Errorf(fmt.Sprintf("set into cache entity id [%v]", res.GetID()), cacheErr)
	}

	return res, err
}

// Delete выполняет удаление (DELETE) записи из базы данных.
// В случае успеха принудительно вырезает (инвалидирует) соответствующий ключ из кэша.
func (bcl *BaseCRUDL2Repository[E, ID]) Delete(ctx context.Context, id ID) error {
	err := bcl.next.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Принудительная инвалидация кэша для соблюдения строгой консистентности данных агрегата
	bcl.crudCache.Delete(id)

	return nil
}

// GetInfo возвращает метаданные сущности EntityInfo (названия таблиц, схем и полей).
func (bcl *BaseCRUDL2Repository[E, ID]) GetInfo() *EntityInfo {
	return bcl.entityInfo
}

// GetCache возвращает ссылку на инжектированный потокобезопасный интерфейс кэш-драйвера.
func (bcl *BaseCRUDL2Repository[E, ID]) GetCache() cache.Cache[ID, E] {
	return bcl.crudCache
}

// GetDefaultTTL возвращает текущее сконфигурированное время жизни (TTL) кэш-записей по умолчанию.
func (bcl *BaseCRUDL2Repository[E, ID]) GetDefaultTTL() time.Duration {
	return bcl.defaultTTL
}

// GetLogger возвращает инстанс изолированного структурированного логгера репозитория.
func (bcl *BaseCRUDL2Repository[E, ID]) GetLogger() logger.Logger {
	return bcl.log
}
