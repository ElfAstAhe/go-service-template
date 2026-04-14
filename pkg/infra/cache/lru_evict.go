package cache

import (
	"container/list"
)

type LRUEvict[K comparable] struct {
	ll    *list.List
	items map[K]*list.Element
}

func NewLRUEvict[K comparable]() *LRUEvict[K] {
	return &LRUEvict[K]{
		ll:    list.New(),
		items: make(map[K]*list.Element),
	}
}

func (lru *LRUEvict[K]) OnGet(key K) {
	if el, ok := lru.items[key]; ok {
		lru.ll.MoveToFront(el)
	}
}

func (lru *LRUEvict[K]) OnSet(key K) {
	if el, ok := lru.items[key]; ok {
		lru.ll.MoveToFront(el)
		return
	}

	lru.items[key] = lru.ll.PushFront(key)
}

func (lru *LRUEvict[K]) OnRemove(key K) {
	if el, ok := lru.items[key]; ok {
		lru.ll.Remove(el)
		delete(lru.items, key)
	}
}

func (lru *LRUEvict[K]) Reset() {
	lru.ll.Init()
	lru.items = make(map[K]*list.Element)
}

func (lru *LRUEvict[K]) Evict() (K, bool) {
	// Самый старый (Least Recently Used) всегда в хвосте
	el := lru.ll.Back()
	if el == nil {
		var zero K
		return zero, false
	}

	key := el.Value.(K)
	lru.OnRemove(key)

	return key, true
}
