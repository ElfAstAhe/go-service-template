package cache

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

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
func Factory[K comparable, V any](options ...FactoryOption[K, V]) (Cache[K, V], error) {
	opts := NewFactoryOptions[K, V]()
	for _, option := range options {
		option(opts)
	}

	if err := opts.Validate(); err != nil {
		return nil, errs.NewCommonError("cache factory options validation failed", err)
	}

	// Страховочный барьер: если стратегия вытеснения не задана — по умолчанию подключаем эталонный LRU
	if utils.IsNil(opts.Policy) {
		opts.Policy = NewLRUEvict[K]()
	}

	// Собираем топологию хранения на основе коэффициента шардирования
	var storage Storage[K]
	switch {
	case opts.ShardCount == 1:
		storage = opts.ShardFactory(opts.MaxSize, opts.Policy)
	case opts.ShardCount > 1:
		storage = NewShardStorage[K](opts.ShardCount, opts.ShardFactory, opts.MaxSize, opts.Policy)
	default:
		return nil, errs.NewCommonError(fmt.Sprintf("invalid shard count [%d]", opts.ShardCount), nil)
	}

	// Если запрошен гибридный L2 режим — оборачиваем хранилище в распределенный сетевой мост
	if opts.L2 {
		return NewL2[K, V](storage, opts.Codec, opts.JanitorMaxSize), nil
	}

	// Возвращаем классический быстрый in-memory кэш-контейнер
	return New[K, V](storage, opts.Codec, opts.JanitorMaxSize), nil
}
