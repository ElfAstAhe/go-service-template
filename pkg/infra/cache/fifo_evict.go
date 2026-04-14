package cache

import (
	"container/list"
	"sync"
)

type FIFOEvict[K comparable] struct {
	mu    sync.RWMutex
	ll    *list.List
	items map[K]*list.Element
}

func NewFIFOEvict[K comparable]() *FIFOEvict[K] {
	return &FIFOEvict[K]{
		ll:    list.New(),
		items: make(map[K]*list.Element),
	}
}

func (fif *FIFOEvict[K]) OnGet(key K) {
	// В FIFO чтение никак не влияет на приоритет вытеснения
}

func (fif *FIFOEvict[K]) OnSet(key K) {
	fif.mu.Lock()
	defer fif.mu.Unlock()

	// Если ключ уже есть, в классическом FIFO его позиция не меняется.
	// Если ключа нет — добавляем в конец очереди (Front или Back — неважно, главное консистентность)
	if _, ok := fif.items[key]; !ok {
		fif.items[key] = fif.ll.PushFront(key)
	}
}

func (fif *FIFOEvict[K]) OnRemove(key K) {
	fif.mu.Lock()
	defer fif.mu.Unlock()

	if el, ok := fif.items[key]; ok {
		fif.ll.Remove(el)
		delete(fif.items, key)
	}
}

func (fif *FIFOEvict[K]) Reset() {
	fif.mu.Lock()
	defer fif.mu.Unlock()

	fif.ll.Init()
	clear(fif.items)
}

func (fif *FIFOEvict[K]) Evict() (K, bool) {
	// Самый "старый" элемент всегда в хвосте (так как новые пушим в Front)
	el := fif.ll.Back()
	if el == nil {
		var zero K
		return zero, false
	}

	key := el.Value.(K)
	fif.OnRemove(key) // Используем OnRemove для очистки индексов

	return key, true
}
