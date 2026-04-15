package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestCacheFactory(t *testing.T) {
	// Общий кодек для тестов
	factory := func() string { return "" }
	jsonCodec := cache.NewJSONCodec[string](factory)

	t.Run("success_simple_cache", func(t *testing.T) {
		c, err := cache.CacheFactory(
			cache.WithCodec[string, string](jsonCodec),
			cache.WithLRUEvictPolicy[string, string](),
		)

		assert.NoError(t, err)
		assert.NotNil(t, c)
		// Проверяем, что это не L2
		_, isL2 := c.(*cache.L2Manager[string, string])
		assert.False(t, isL2)
	})

	t.Run("success_l2_sharded_cache", func(t *testing.T) {
		c, err := cache.CacheFactory(
			cache.WithCodec[string, string](jsonCodec),
			cache.WithL2Cache[string, string](),
			cache.WithShardCount[string, string](16),
			cache.WithLFUEvictPolicy[string, string](),
		)

		assert.NoError(t, err)
		// Проверяем, что создался именно L2Manager
		_, isL2 := c.(*cache.L2Manager[string, string])
		assert.True(t, isL2)
	})

	t.Run("error_missing_codec", func(t *testing.T) {
		// Не передаем кодек
		c, err := cache.CacheFactory[string, string](
			cache.WithLRUEvictPolicy[string, string](),
		)

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "cache codec not applied")
	})

	t.Run("error_invalid_shard_count", func(t *testing.T) {
		_, err := cache.CacheFactory(
			cache.WithCodec[string, string](jsonCodec),
			cache.WithShardCount[string, string](0), // Инвалидный каунт
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache shard count must be greater zero")
	})

	t.Run("check_default_policy", func(t *testing.T) {
		// Не передаем политику, должна поставиться LRU по умолчанию (внутри фабрики)
		c, err := cache.CacheFactory(
			cache.WithCodec[string, string](jsonCodec),
		)

		assert.NoError(t, err)
		assert.NotNil(t, c)
		// Проверяем работоспособность
		err = c.Set("test", "data", 0)
		assert.NoError(t, err)
	})
}
