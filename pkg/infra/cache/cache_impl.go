package cache

import (
	"time"
)

// cache — основная реализация интерфейса Cache[K, V]
type cache[K comparable, V any] struct {
	storage Storage[K]
	codec   Codec[V]
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
		var zero V
		return zero, false, nil
	}

	env, err := c.codec.Unmarshal(b)
	if err != nil {
		var zero V
		return zero, false, err
	}

	// Проверка TTL (ленивое удаление)
	if env.DieAt > 0 && time.Now().UnixNano() > env.DieAt {
		c.storage.Delete(key)
		var zero V
		return zero, false, nil
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
func (c *cache[K, V]) CacheJanitor(eventTime time.Time) error {
	now := eventTime.UnixNano()
	var expiredKeys []K

	// Собираем ключи для удаления, чтобы не блокировать storage надолго
	c.storage.Range(func(key K, b []byte) bool {
		env, err := c.codec.Unmarshal(b)
		if err == nil && env.DieAt > 0 && now > env.DieAt {
			expiredKeys = append(expiredKeys, key)
		}
		return true
	})

	for _, k := range expiredKeys {
		c.storage.Delete(k)
	}

	return nil
}
