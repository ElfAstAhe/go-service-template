package cache

import (
	"container/list"
	"sync"
)

type LFUItem[K comparable] struct {
	key   K
	count int
}

type LFUEvict[K comparable] struct {
	mu       sync.Mutex // Только Lock, только победа
	items    map[K]*list.Element
	freqs    map[int]*list.List
	minFreq  int
	emptyKey K
}

func NewLFUEvict[K comparable]() *LFUEvict[K] {
	return &LFUEvict[K]{
		items: make(map[K]*list.Element),
		freqs: make(map[int]*list.List),
	}
}

func (lfu *LFUEvict[K]) OnGet(key K) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	el, ok := lfu.items[key]
	if !ok {
		return
	}

	lfu.increment(el)
}

func (lfu *LFUEvict[K]) OnSet(key K) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if el, ok := lfu.items[key]; ok {
		lfu.increment(el)
		return
	}

	// Новый элемент
	item := &LFUItem[K]{key: key, count: 1}
	lfu.minFreq = 1
	lfu.insert(item)
}

func (lfu *LFUEvict[K]) OnRemove(key K) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.remove(key)
}

func (lfu *LFUEvict[K]) Evict() (K, bool) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if len(lfu.items) == 0 {
		return lfu.emptyKey, false
	}

	lst := lfu.freqs[lfu.minFreq]
	el := lst.Back()
	if el == nil {
		return lfu.emptyKey, false
	}

	key := el.Value.(*LFUItem[K]).key
	lfu.remove(key) // Используем внутренний неблокирующий метод

	return key, true
}

func (lfu *LFUEvict[K]) Reset() {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.items = make(map[K]*list.Element)
	lfu.freqs = make(map[int]*list.List)
	lfu.minFreq = 0
}

// --- Внутренние неблокирующие методы (Helper methods) ---

func (lfu *LFUEvict[K]) increment(el *list.Element) {
	item := el.Value.(*LFUItem[K])
	oldFreq := item.count

	// Убираем из старого списка частот
	lfu.freqs[oldFreq].Remove(el)

	// Обновляем minFreq если нужно
	if lfu.freqs[oldFreq].Len() == 0 && oldFreq == lfu.minFreq {
		lfu.minFreq++
	}

	item.count++
	lfu.insert(item)
}

func (lfu *LFUEvict[K]) insert(item *LFUItem[K]) {
	if _, ok := lfu.freqs[item.count]; !ok {
		lfu.freqs[item.count] = list.New()
	}
	el := lfu.freqs[item.count].PushFront(item)
	lfu.items[item.key] = el
}

func (lfu *LFUEvict[K]) remove(key K) {
	if el, ok := lfu.items[key]; ok {
		item := el.Value.(*LFUItem[K])
		lfu.freqs[item.count].Remove(el)
		delete(lfu.items, key)
	}
}
