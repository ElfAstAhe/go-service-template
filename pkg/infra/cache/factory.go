package cache

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type factoryConfig[K comparable, V any] struct {
	l2             bool
	shardCount     uint64
	shardFactory   ShardFactory[K]
	maxSize        int
	policy         EvictionPolicy[K]
	codec          Codec[V]
	janitorMaxSize int
}

func (fc *factoryConfig[K, V]) Validate() error {
	// max size
	if fc.maxSize < 0 {
		return errs.NewCommonError("max cache size must be greater or equal zero", nil)
	}
	// shard count
	if fc.shardCount == 0 {
		return errs.NewCommonError("cache shard count must be greater zero", nil)
	}
	// shard factory
	if utils.IsNil(fc.shardFactory) {
		return errs.NewCommonError("cache shard factory must be applied", nil)
	}
	// codec
	if utils.IsNil(fc.codec) {
		return errs.NewCommonError("cache codec not applied", nil)
	}
	// janitor max capacity
	if fc.janitorMaxSize <= 0 {
		return errs.NewCommonError("janitor max size must be greater than zero", nil)
	}

	return nil
}

type Option[K comparable, V any] func(config *factoryConfig[K, V])

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

func CacheFactory[K comparable, V any](opts ...Option[K, V]) (Cache[K, V], error) {
	// default config
	conf := defaultFactoryConfig[K, V]()

	// options
	for _, opt := range opts {
		opt(conf)
	}

	// validate
	if err := conf.Validate(); err != nil {
		return nil, errs.NewCommonError("cache factory invalid config", err)
	}

	// check policy
	if utils.IsNil(conf.policy) {
		conf.policy = NewLRUEvict[K]()
	}

	// storage
	var storage Storage[K]
	switch {
	case conf.shardCount == 1:
		storage = conf.shardFactory(conf.maxSize, conf.policy)
	case conf.shardCount > 1:
		storage = NewShardStorage[K](conf.shardCount, conf.shardFactory, conf.maxSize, conf.policy)
	default:
		return nil, errs.NewCommonError(fmt.Sprintf("invalid shard count [%d]", conf.shardCount), nil)
	}

	// L2 cache
	if conf.l2 {
		return NewL2[K, V](storage, conf.codec, conf.janitorMaxSize), nil
	}

	// cache
	return New[K, V](storage, conf.codec, conf.janitorMaxSize), nil
}

func WithL2Cache[K comparable, V any]() Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.l2 = true
	}
}

func WithShardCount[K comparable, V any](shardCount uint64) Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.shardCount = shardCount
	}
}

func WithShardFactory[K comparable, V any](shardFactory ShardFactory[K]) Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.shardFactory = shardFactory
	}
}

func WithMaxSize[K comparable, V any](maxSize int) Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.maxSize = maxSize
	}
}

func WithLRUEvictPolicy[K comparable, V any]() Option[K, V] {
	return WithCustomEvictPolicy[K, V](NewLRUEvict[K]())
}

func WithLFUEvictPolicy[K comparable, V any]() Option[K, V] {
	return WithCustomEvictPolicy[K, V](NewLFUEvict[K]())
}

func WithFIFOEvictPolicy[K comparable, V any]() Option[K, V] {
	return WithCustomEvictPolicy[K, V](NewFIFOEvict[K]())
}

func WithCustomEvictPolicy[K comparable, V any](policy EvictionPolicy[K]) Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.policy = policy
	}
}

func WithCodec[K comparable, V any](codec Codec[V]) Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.codec = codec
	}
}

func WithJanitorMaxSize[K comparable, V any](janitorMaxSize int) Option[K, V] {
	return func(config *factoryConfig[K, V]) {
		config.janitorMaxSize = janitorMaxSize
	}
}
