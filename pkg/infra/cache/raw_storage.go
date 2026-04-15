package cache

import (
	"sync"
)

type RawStorage[K comparable] struct {
	mu      sync.RWMutex
	data    map[K][]byte
	policy  EvictionPolicy[K]
	maxSize int
}

func NewRawStorage[K comparable](maxSize int, policy EvictionPolicy[K]) *RawStorage[K] {
	return &RawStorage[K]{
		data:    make(map[K][]byte),
		maxSize: maxSize,
		policy:  policy,
	}
}

func (rs *RawStorage[K]) Get(key K) ([]byte, bool) {
	rs.mu.Lock() // Lock, так как OnGet в LRU/LFU меняет состояние (двигает элементы)
	defer rs.mu.Unlock()

	b, ok := rs.data[key]
	if ok {
		rs.policy.OnGet(key)
	}

	return b, ok
}

func (rs *RawStorage[K]) Set(key K, b []byte) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Если ключа нет, проверяем, не пора ли кого-то выселить
	if _, exists := rs.data[key]; !exists {
		if rs.maxSize > 0 && len(rs.data) >= rs.maxSize {
			if victim, ok := rs.policy.Evict(); ok {
				delete(rs.data, victim)
			}
		}
	}

	rs.data[key] = b
	rs.policy.OnSet(key)
}

func (rs *RawStorage[K]) Delete(key K) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	delete(rs.data, key)
	rs.policy.OnRemove(key)
}

func (rs *RawStorage[K]) Range(fn func(key K, value []byte) bool) {
	// Используем RLock, чтобы разрешить параллельное чтение (Get),
	// но заблокировать запись (Set/Delete) на время итерации.
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	for k, v := range rs.data {
		if !fn(k, v) {
			break
		}
	}
}

// Has проверяет наличие ключа без влияния на политику вытеснения
func (rs *RawStorage[K]) Has(key K) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, ok := rs.data[key]
	return ok
}

func (rs *RawStorage[K]) Len() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.data)
}

func (rs *RawStorage[K]) Clear() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.data = make(map[K][]byte)
	rs.policy.Reset()
}
