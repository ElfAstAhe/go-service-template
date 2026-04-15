package cache

import (
	"time"
)

const (
	negativeGetTTL time.Duration = 2 * time.Minute
)

// L2Manager - cache manager с поведением L2, умеет запоминать негативный поиск с добавлением пустых значений
type L2Manager[K comparable, V any] struct {
	*Manager[K, V]
}

func NewL2[K comparable, V any](
	storage Storage[K],
	codec Codec[V],
	janitorMaxSize int,
) *L2Manager[K, V] {
	return &L2Manager[K, V]{
		Manager: New(storage, codec, janitorMaxSize),
	}
}

func (l2m *L2Manager[K, V]) Get(key K) (V, bool, error) {
	res, ok, err := l2m.Manager.Get(key)
	if err != nil {
		return res, ok, err
	}
	if !ok {
		_ = l2m.Manager.Set(key, res, negativeGetTTL)
	}

	return res, ok, nil
}
