package utils

import (
	"iter"
	"sync"
)

type ConcurrentList[T any] struct {
	mu    sync.RWMutex
	items []T
}

func NewConcurrentList[T any]() *ConcurrentList[T] {
	return &ConcurrentList[T]{
		items: make([]T, 0),
	}
}

// Append добавляет элемент в конец списка
func (l *ConcurrentList[T]) Append(item T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, item)
}

// Get возвращает элемент по индексу
func (l *ConcurrentList[T]) Get(index int) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var zero T
	if index < 0 || index >= len(l.items) {
		return zero, false
	}
	return l.items[index], true
}

// Len возвращает текущую длину списка
func (l *ConcurrentList[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.items)
}

// Snapshot возвращает копию слайса для безопасной итерации
func (l *ConcurrentList[T]) Snapshot() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	cp := make([]T, len(l.items))
	copy(cp, l.items)
	return cp
}

// All возвращает итератор iter.Seq2, отдающий пары [индекс, значение].
// Позволяет писать: for idx, val := range list.All() { ... }
func (l *ConcurrentList[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		// 1. Делаем быстрый снимок данных под RLock, чтобы минимизировать время блокировки.
		l.mu.RLock()
		snapshot := make([]T, len(l.items))
		copy(snapshot, l.items)
		l.mu.RUnlock()

		// 2. Итерируемся по изолированной копии.
		// Горутины-писатели могут спокойно делать Append во время выполнения этого цикла.
		for idx, val := range snapshot {
			// yield возвращает false, если пользователь вызвал break внутри цикла for range
			if !yield(idx, val) {
				return
			}
		}
	}
}

// Values возвращает итератор iter.Seq, отдающий только значения.
// Позволяет писать: for val := range list.Values() { ... }
func (l *ConcurrentList[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		l.mu.RLock()
		snapshot := make([]T, len(l.items))
		copy(snapshot, l.items)
		l.mu.RUnlock()

		for _, val := range snapshot {
			if !yield(val) {
				return
			}
		}
	}
}
