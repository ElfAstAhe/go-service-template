package cache

import (
	"container/list"
	"sync"
)

type LRUEvict[K comparable] struct {
	mu       sync.Mutex
	ll       *list.List
	items    map[K]*list.Element
	emptyKey K
}

func NewLRUEvict[K comparable]() *LRUEvict[K] {
	return &LRUEvict[K]{
		ll:    list.New(),
		items: make(map[K]*list.Element),
	}
}

func (lru *LRUEvict[K]) OnGet(key K) {
	lru.mu.Lock() // Только Lock!
	defer lru.mu.Unlock()
	if el, ok := lru.items[key]; ok {
		lru.ll.MoveToFront(el)
	}
}

func (lru *LRUEvict[K]) OnSet(key K) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// Если ключ уже есть, просто переносим в начало очереди (омолаживаем)
	if el, ok := lru.items[key]; ok {
		lru.ll.MoveToFront(el)
		return
	}

	// Если ключа нет, добавляем в начало
	lru.items[key] = lru.ll.PushFront(key)
}

func (lru *LRUEvict[K]) OnRemove(key K) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	lru.remove(key) // Вызов приватного метода
}

func (lru *LRUEvict[K]) Reset() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.ll.Init()
	lru.items = make(map[K]*list.Element)
}

func (lru *LRUEvict[K]) Evict() (K, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	el := lru.ll.Back()
	if el == nil {
		return lru.emptyKey, false
	}

	key := el.Value.(K)
	lru.remove(key) // Теперь дедлока нет, т.к. remove не берет лок
	return key, true
}

// Приватный метод для внутренней логики без мьютексов
func (lru *LRUEvict[K]) remove(key K) {
	if el, ok := lru.items[key]; ok {
		lru.ll.Remove(el)
		delete(lru.items, key)
	}
}
