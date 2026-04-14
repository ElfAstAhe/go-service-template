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

func (cm *Manager[K, V]) Get(key K) (V, bool, error) {
	buf, ok := cm.storage.Get(key)
	if !ok {
		return cm.nilValue, false, nil
	}

	envelope, err := cm.codec.Unmarshal(buf)
	if err != nil {
		return cm.nilValue, false, errs.NewCommonError("unmarshal failed", err)
	}

	// Проверка TTL (ленивое удаление)
	if envelope.DieAt > 0 && time.Now().UnixNano() > envelope.DieAt {
		cm.storage.Delete(key)
		return cm.nilValue, false, nil
	}

	return envelope.Value, true, nil
}

func (cm *Manager[K, V]) Set(key K, value V, ttl time.Duration) error {
	buf, err := cm.codec.Marshal(value, ttl)
	if err != nil {
		return errs.NewCommonError("marshal failed", err)
	}

	cm.storage.Set(key, buf)

	return nil
}

func (cm *Manager[K, V]) Delete(key K) {
	cm.storage.Delete(key)
}

func (cm *Manager[K, V]) Size() int {
	return cm.storage.Len()
}

func (cm *Manager[K, V]) Clear() {
	cm.storage.Clear()
}

// CacheJanitor вызывается планировщиком для периодической очистки просрочки
func (cm *Manager[K, V]) CacheJanitor(ctx context.Context, eventTime time.Time) error {
	now := eventTime.UnixNano()
	var expiredKeys []K
	if cm.janitorMaxSize > 0 {
		expiredKeys = make([]K, 0, cm.janitorMaxSize)
	} else {
		expiredKeys = make([]K, 0)
	}

	var janitorCount int
	// Собираем ключи для удаления, чтобы не блокировать storage надолго
	cm.storage.Range(func(key K, b []byte) bool {
		// check for janitor max size
		if cm.janitorMaxSize > 0 && janitorCount >= cm.janitorMaxSize {
			return false
		}
		// check context
		if err := ctx.Err(); err != nil {
			return false
		}
		// unmarshal
		env, err := cm.codec.Unmarshal(b)
		// check for removal and add into removal list
		if err == nil && env != nil && env.DieAt > 0 && now > env.DieAt {
			expiredKeys = append(expiredKeys, key)
			janitorCount++
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
			cm.storage.Delete(k)
		}
	}

	return nil
}
