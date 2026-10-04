package cache

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const (
	DefaultCacheL2             bool   = false
	DefaultCacheShardCount     uint64 = 1
	DefaultCacheMaxSize        int    = 10000
	DefaultCacheJanitorMaxSize int    = 1000
)

// FactoryOption определяет функциональный тип конфигуратора (Fluent API) для сборки кэш-системы.
type FactoryOption[K comparable, V any] func(config *FactoryOptions[K, V])

// FactoryOptions инкапсулирует внутренние параметры рантайма, необходимые для сборки
// и пре-аллокации памяти высокопроизводительного дженерик-кэша.
type FactoryOptions[K comparable, V any] struct {
	L2             bool              // Флаг активации двухуровневого кэширования (L1 Local + L2 Distributed)
	ShardCount     uint64            // Количество изолированных шардов памяти для снижения Mutex Contention
	ShardFactory   ShardFactory[K]   // Фабричный метод для сборки сегментов хранилища Storage
	MaxSize        int               // Максимально допустимый лимит элементов в кэше (DataCapacity)
	Policy         EvictionPolicy[K] // Выбранный контроллер стратегии вытеснения данных (LRU, LFU, FIFO)
	Codec          Codec[V]          // Кодек сериализации конвертов для L2/сетевого взаимодействия
	JanitorMaxSize int               // Верхний потолок емкости фонового буфера очистки протухших TTL-записей
}

// NewFactoryOptions наполняет конфиг безопасными дефолтами: 1 шард, емкость 10k объектов и стратегия RawStorage.
func NewFactoryOptions[K comparable, V any]() *FactoryOptions[K, V] {
	return &FactoryOptions[K, V]{
		L2:         DefaultCacheL2,
		ShardCount: DefaultCacheShardCount,
		ShardFactory: func(maxSize int, policy EvictionPolicy[K]) Storage[K] {
			return NewRawStorage[K](maxSize, policy)
		},
		MaxSize:        DefaultCacheMaxSize,
		JanitorMaxSize: DefaultCacheJanitorMaxSize,
	}
}

// Validate выполняет строгую проверку параметров конфигурации перед аллокацией памяти кэша,
// защищая приложение от логических ошибок планировщика и паник из-за nil-зависимостей.
func (fc *FactoryOptions[K, V]) Validate() error {
	if fc.ShardCount <= 0 {
		return errs.NewCommonError("cache shard count must be greater zero", nil)
	}
	if utils.IsNil(fc.ShardFactory) {
		return errs.NewCommonError("cache shard factory must be applied", nil)
	}
	if fc.MaxSize < 0 {
		return errs.NewCommonError("max cache size must be greater or equal zero", nil)
	}
	if utils.IsNil(fc.Codec) {
		return errs.NewCommonError("cache codec not applied", nil)
	}
	if fc.JanitorMaxSize <= 0 {
		return errs.NewCommonError("janitor max size must be greater than zero", nil)
	}

	return nil
}

// ====================================================================
// Набор Fluent API опций конфигурации (Опциональные мутаторы)
// ====================================================================

// WithCacheL2 переводит подсистему в режим гибридного распределенного двухуровневого кэша.
func WithCacheL2[K comparable, V any](l2 bool) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) {
		config.L2 = l2
	}
}

// WithCacheShardCount задает количество независимых сегментов памяти во избежание Mutex Contention блокировок.
func WithCacheShardCount[K comparable, V any](shardCount uint64) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) {
		config.ShardCount = shardCount
	}
}

// WithCacheShardFactory выполняет инжекцию кастомного фабричного билдера для низкоуровневых шардов.
func WithCacheShardFactory[K comparable, V any](f ShardFactory[K]) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) {
		config.ShardFactory = f
	}
}

// WithCacheMaxSize ограничивает верхний потолок емкости (количество одновременно удерживаемых ключей).
func WithCacheMaxSize[K comparable, V any](size int) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) {
		config.MaxSize = size
	}
}

// WithCacheLRUEvictPolicy активирует стратегию вытеснения наименее недавно использовавшихся элементов (Least Recently Used).
func WithCacheLRUEvictPolicy[K comparable, V any]() FactoryOption[K, V] {
	return WithCacheCustomEvictPolicy[K, V](NewLRUEvict[K]())
}

// WithCacheLFUEvictPolicy активирует стратегию вытеснения наименее часто использовавшихся элементов (Least Frequently Used).
func WithCacheLFUEvictPolicy[K comparable, V any]() FactoryOption[K, V] {
	return WithCacheCustomEvictPolicy[K, V](NewLFUEvict[K]())
}

// WithCacheFIFOEvictPolicy активирует последовательную стратегию вытеснения в порядке очереди (First In, First Out).
func WithCacheFIFOEvictPolicy[K comparable, V any]() FactoryOption[K, V] {
	return WithCacheCustomEvictPolicy[K, V](NewFIFOEvict[K]())
}

// WithCacheCustomEvictPolicy инжектирует кастомную/внешнюю математическую стратегию вытеснения ключей.
func WithCacheCustomEvictPolicy[K comparable, V any](policy EvictionPolicy[K]) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) { config.Policy = policy }
}

// WithCacheCodec регистрирует маршалер для сериализации/десериализации бинарных пакетов.
func WithCacheCodec[K comparable, V any](codec Codec[V]) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) {
		config.Codec = codec
	}
}

// WithCacheJanitorMaxSize настраивает емкость фонового планировщика очистки TTL-мусора в оперативной памяти.
func WithCacheJanitorMaxSize[K comparable, V any](janitorMaxSize int) FactoryOption[K, V] {
	return func(config *FactoryOptions[K, V]) {
		config.JanitorMaxSize = janitorMaxSize
	}
}
