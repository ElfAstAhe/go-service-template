package cache

import (
	"time"
)

// Cache — универсальный интерфейс кэша данных с generics
type Cache[K comparable, V any] interface {
	// Get возвращает десериализованную копию объекта.
	Get(key K) (V, bool, error)

	// Set сериализует объект и сохраняет его в кэш.
	Set(key K, value V, ttl time.Duration) error

	// Delete удаляет ключ из кэша.
	Delete(key K)

	// Служебные методы
	Size() int
	Clear()
}
