package cache

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Manager — основная реализация интерфейса Cache[K, V]
type Manager[K comparable, V any] struct {
	// storage Хранилище данных кэша
	storage Storage[K]
	// codec упаковка и распаковка
	codec Codec[V]
	// nilValue пустое значение
	nilValue V
	// janitorMaxSize максимальное кол-во элементов на удаление по TTL
	janitorMaxSize int
}

// New создает новый экземпляр кэша
func New[K comparable, V any](
	storage Storage[K],
	codec Codec[V],
	janitorMaxSize int,
) *Manager[K, V] {
	return &Manager[K, V]{
		storage:        storage,
		codec:          codec,
		janitorMaxSize: janitorMaxSize,
	}
}

func (c *Manager[K, V]) Get(key K) (V, bool, error) {
	buf, ok := c.storage.Get(key)
	if !ok {
		return c.nilValue, false, nil
	}

	envelope, err := c.codec.Unmarshal(buf)
	if err != nil {
		return c.nilValue, false, errs.NewCommonError("unmarshal failed", err)
	}

	// Проверка TTL (ленивое удаление)
	if envelope.DieAt > 0 && time.Now().UnixNano() > envelope.DieAt {
		c.storage.Delete(key)
		return c.nilValue, false, nil
	}

	return envelope.Value, true, nil
}

func (c *Manager[K, V]) Set(key K, value V, ttl time.Duration) error {
	buf, err := c.codec.Marshal(value, ttl)
	if err != nil {
		return errs.NewCommonError("marshal failed", err)
	}

	c.storage.Set(key, buf)

	return nil
}

func (c *Manager[K, V]) Delete(key K) {
	c.storage.Delete(key)
}

func (c *Manager[K, V]) Size() int {
	return c.storage.Len()
}

func (c *Manager[K, V]) Clear() {
	c.storage.Clear()
}

// CacheJanitor вызывается планировщиком для периодической очистки просрочки
func (c *Manager[K, V]) CacheJanitor(ctx context.Context, eventTime time.Time) error {
	now := eventTime.UnixNano()
	var expiredKeys []K

	var janitorCount int
	// Собираем ключи для удаления, чтобы не блокировать storage надолго
	c.storage.Range(func(key K, b []byte) bool {
		// unmarshal
		env, err := c.codec.Unmarshal(b)
		// check for removal and add into removal list
		if err == nil && env.DieAt > 0 && now > env.DieAt {
			expiredKeys = append(expiredKeys, key)
			janitorCount++
		}
		// check for janitor max size
		if c.janitorMaxSize > 0 && janitorCount >= c.janitorMaxSize {
			return false
		}
		// check context
		if err := ctx.Err(); err != nil {
			return false
		}

		// approve next iteration
		return true
	})

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// removal
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
