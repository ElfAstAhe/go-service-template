package cache

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// factoryConfig инкапсулирует внутренние параметры рантайма, необходимые для сборки
// и пре-аллокации памяти высокопроизводительного дженерик-кэша.
type factoryConfig[K comparable, V any] struct {
	l2             bool              // Флаг активации двухуровневого кэширования (L1 Local + L2 Distributed)
	shardCount     uint64            // Количество изолированных шардов памяти для снижения Mutex Contention
	shardFactory   ShardFactory[K]   // Фабричный метод для сборки сегментов хранилища Storage
	maxSize        int               // Максимально допустимый лимит элементов в кэше (DataCapacity)
	policy         EvictionPolicy[K] // Выбранный контроллер стратегии вытеснения данных (LRU, LFU, FIFO)
	codec          Codec[V]          // Кодек сериализации конвертов для L2/сетевого взаимодействия
	janitorMaxSize int               // Верхний потолок емкости фонового буфера очистки протухших TTL-записей
}

// Validate выполняет строгую проверку параметров конфигурации перед аллокацией памяти кэша,
// защищая приложение от логических ошибок планировщика и паник из-за nil-зависимостей.
func (fc *factoryConfig[K, V]) Validate() error {
	if fc.maxSize < 0 {
		return errs.NewCommonError("max cache size must be greater or equal zero", nil)
	}
	if fc.shardCount == 0 {
		return errs.NewCommonError("cache shard count must be greater zero", nil)
	}
	if utils.IsNil(fc.shardFactory) {
		return errs.NewCommonError("cache shard factory must be applied", nil)
	}
	if utils.IsNil(fc.codec) {
		return errs.NewCommonError("cache codec not applied", nil)
	}
	if fc.janitorMaxSize <= 0 {
		return errs.NewCommonError("janitor max size must be greater than zero", nil)
	}
	return nil
}

// Option определяет функциональный тип конфигуратора (Fluent API) для сборки кэш-системы.
type Option[K comparable, V any] func(config *factoryConfig[K, V])

// defaultFactoryConfig наполняет конфиг безопасными дефолтами: 1 шард, емкость 10k объектов и стратегия RawStorage.
func defaultFactoryConfig[K comparable, V any]() *factoryConfig[K, V] {
	return &factoryConfig[K, V]{
		l2:         false,
		shardCount: 1,
		shardFactory: func(maxSize int, policy EvictionPolicy[K]) Storage[K] {
			return NewRawStorage[K](maxSize, policy)
		},
		maxSize:        10000,
		janitorMaxSize: 1000,
	}
}

// Factory — центральный декларативный конструктор (Abstract Factory) подсистемы кэширования фреймворка.
//
// На основе переданных функциональных опций автоматически собирает нужную топологию рантайма:
// - Конкурентный однопоточный кэш (shardCount == 1)
// - Высококонкурентный шардированный сегментированный кэш (shardCount > 1)
// - Двухуровневый гибридный кэш (with L2 Distributed Storage)
//
//goland:noinspection GoNameStartsWithPackageName
//lint:ignore exported This name is kept for backward compatibility and domain clarity
//revive:ignore
func Factory[K comparable, V any](opts ...Option[K, V]) (Cache[K, V], error) {
	conf := defaultFactoryConfig[K, V]()

	// Вычисляем входящие мутаторы Fluent API
	for _, opt := range opts {
		opt(conf)
	}

	if err := conf.Validate(); err != nil {
		return nil, errs.NewCommonError("cache factory invalid config", err)
	}

	// Страховочный барьер: если стратегия вытеснения не задана — по умолчанию подключаем эталонный LRU
	if utils.IsNil(conf.policy) {
		conf.policy = NewLRUEvict[K]()
	}

	// Собираем топологию хранения на основе коэффициента шардирования
	var storage Storage[K]
	switch {
	case conf.shardCount == 1:
		storage = conf.shardFactory(conf.maxSize, conf.policy)
	case conf.shardCount > 1:
		storage = NewShardStorage[K](conf.shardCount, conf.shardFactory, conf.maxSize, conf.policy)
	default:
		return nil, errs.NewCommonError(fmt.Sprintf("invalid shard count [%d]", conf.shardCount), nil)
	}

	// Если запрошен гибридный L2 режим — оборачиваем хранилище в распределенный сетевой мост
	if conf.l2 {
		return NewL2[K, V](storage, conf.codec, conf.janitorMaxSize), nil
	}

	// Возвращаем классический быстрый in-memory кэш-контейнер
	return New[K, V](storage, conf.codec, conf.janitorMaxSize), nil
}

// ====================================================================
// Набор Fluent API опций конфигурации (Опциональные мутаторы)
// ====================================================================

// WithL2Cache переводит подсистему в режим гибридного распределенного двухуровневого кэша.
func WithL2Cache[K comparable, V any]() Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.l2 = true }
}

// WithShardCount задает количество независимых сегментов памяти во избежание Mutex Contention блокировок.
func WithShardCount[K comparable, V any](shardCount uint64) Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.shardCount = shardCount }
}

// WithShardFactory выполняет инжекцию кастомного фабричного билдера для низкоуровневых шардов.
func WithShardFactory[K comparable, V any](shardFactory ShardFactory[K]) Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.shardFactory = shardFactory }
}

// WithMaxSize ограничивает верхний потолок емкости (количество одновременно удерживаемых ключей).
func WithMaxSize[K comparable, V any](maxSize int) Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.maxSize = maxSize }
}

// WithLRUEvictPolicy активирует стратегию вытеснения наименее недавно использовавшихся элементов (Least Recently Used).
func WithLRUEvictPolicy[K comparable, V any]() Option[K, V] {
	return WithCustomEvictPolicy[K, V](NewLRUEvict[K]())
}

// WithLFUEvictPolicy активирует стратегию вытеснения наименее часто использовавшихся элементов (Least Frequently Used).
func WithLFUEvictPolicy[K comparable, V any]() Option[K, V] {
	return WithCustomEvictPolicy[K, V](NewLFUEvict[K]())
}

// WithFIFOEvictPolicy активирует последовательную стратегию вытеснения в порядке очереди (First In, First Out).
func WithFIFOEvictPolicy[K comparable, V any]() Option[K, V] {
	return WithCustomEvictPolicy[K, V](NewFIFOEvict[K]())
}

// WithCustomEvictPolicy инжектирует кастомную/внешнюю математическую стратегию вытеснения ключей.
func WithCustomEvictPolicy[K comparable, V any](policy EvictionPolicy[K]) Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.policy = policy }
}

// WithCodec регистрирует маршалер для сериализации/десериализации бинарных пакетов.
func WithCodec[K comparable, V any](codec Codec[V]) Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.codec = codec }
}

// WithJanitorMaxSize настраивает емкость фонового планировщика очистки TTL-мусора в оперативной памяти.
func WithJanitorMaxSize[K comparable, V any](janitorMaxSize int) Option[K, V] {
	return func(config *factoryConfig[K, V]) { config.janitorMaxSize = janitorMaxSize }
}
