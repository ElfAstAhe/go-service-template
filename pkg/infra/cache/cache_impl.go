package cache

import (
	"context"
	"time"
)

// cache — основная реализация интерфейса Cache[K, V]
type cache[K comparable, V any] struct {
	storage  Storage[K]
	codec    Codec[V]
	nilValue V
}

// New создает новый экземпляр кэша и запускает фоновую очистку через планировщик
func New[K comparable, V any](
	storage Storage[K],
	codec Codec[V],
) Cache[K, V] {
	return &cache[K, V]{
		storage: storage,
		codec:   codec,
	}
}

func (c *cache[K, V]) Get(key K) (V, bool, error) {
	b, ok := c.storage.Get(key)
	if !ok {
		return c.nilValue, false, nil
	}

	env, err := c.codec.Unmarshal(b)
	if err != nil {
		return c.nilValue, false, err
	}

	// Проверка TTL (ленивое удаление)
	if env.DieAt > 0 && time.Now().UnixNano() > env.DieAt {
		c.storage.Delete(key)
		return c.nilValue, false, nil
	}

	return env.Value, true, nil
}

func (c *cache[K, V]) Set(key K, value V, ttl time.Duration) error {
	b, err := c.codec.Marshal(value, ttl)
	if err != nil {
		return err
	}

	c.storage.Set(key, b)

	return nil
}

func (c *cache[K, V]) Delete(key K) {
	c.storage.Delete(key)
}

func (c *cache[K, V]) Size() int {
	return c.storage.Len()
}

func (c *cache[K, V]) Clear() {
	c.storage.Clear()
}

// CacheJanitor вызывается планировщиком для периодической очистки просрочки
func (c *cache[K, V]) CacheJanitor(ctx context.Context, eventTime time.Time) error {
	now := eventTime.UnixNano()
	var expiredKeys []K

	// Собираем ключи для удаления, чтобы не блокировать storage надолго
	c.storage.Range(func(key K, b []byte) bool {
		// unmarshal
		env, err := c.codec.Unmarshal(b)
		// check for removal and add into removal list
		if err == nil && env.DieAt > 0 && now > env.DieAt {
			expiredKeys = append(expiredKeys, key)
		}
		// check context
		if err := ctx.Err(); err != nil {
			return false
		}

		// approve next iteration
		return true
	})

	for _, k := range expiredKeys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			c.storage.Delete(k)
		}
	}

	return nil
}
