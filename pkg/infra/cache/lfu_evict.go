package cache

import (
	"container/list"
)

type LFUItem[K comparable] struct {
	key   K
	count int // Счетчик частоты
}

type LFUEvict[K comparable] struct {
	items   map[K]*list.Element // Быстрый доступ к узлу
	freqs   map[int]*list.List  // Списки по частотам: 1 -> [keyA, keyB], 2 -> [keyC]
	minFreq int                 // Самая низкая частота на данный момент
}

func NewLFUEvict[K comparable]() *LFUEvict[K] {
	return &LFUEvict[K]{
		items:   make(map[K]*list.Element),
		freqs:   make(map[int]*list.List),
		minFreq: 0,
	}
}

func (lfu *LFUEvict[K]) OnGet(key K) {
	el, ok := lfu.items[key]
	if !ok {
		return
	}

	item := el.Value.(*LFUItem[K])
	oldFreq := item.count
	item.count++

	// 1. Убираем из текущего списка частот
	lfu.freqs[oldFreq].Remove(el)

	// 2. Если текущий список стал пустым и это была минимальная частота — двигаем minFreq
	if lfu.freqs[oldFreq].Len() == 0 && oldFreq == lfu.minFreq {
		lfu.minFreq++
	}

	// 3. Переносим в новый список частот
	lfu.move(item)
}

func (lfu *LFUEvict[K]) OnSet(key K) {
	if _, ok := lfu.items[key]; ok {
		lfu.OnGet(key) // Если обновление — просто инкрементим частоту
		return
	}

	// Новый элемент: всегда частота 1
	item := &LFUItem[K]{key: key, count: 1}
	lfu.minFreq = 1
	lfu.move(item)
}

func (lfu *LFUEvict[K]) OnRemove(key K) {
	if el, ok := lfu.items[key]; ok {
		item := el.Value.(*LFUItem[K])
		lfu.freqs[item.count].Remove(el)
		delete(lfu.items, key)
	}
}

func (lfu *LFUEvict[K]) Reset() {
	lfu.items = make(map[K]*list.Element)
	lfu.freqs = make(map[int]*list.List)
	lfu.minFreq = 0
}

func (lfu *LFUEvict[K]) Evict() (K, bool) {
	if len(lfu.items) == 0 {
		var zero K
		return zero, false
	}

	// Забираем любого "смертника" с минимальной частотой (из хвоста списка)
	list := lfu.freqs[lfu.minFreq]
	el := list.Back()
	key := el.Value.(*LFUItem[K]).key

	lfu.OnRemove(key)
	return key, true
}

// Вспомогательный метод для вставки в нужный список частот
func (lfu *LFUEvict[K]) move(item *LFUItem[K]) {
	if _, ok := lfu.freqs[item.count]; !ok {
		lfu.freqs[item.count] = list.New()
	}
	el := lfu.freqs[item.count].PushFront(item)
	lfu.items[item.key] = el
}
