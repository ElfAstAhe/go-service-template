package utils

import (
	"iter"
	"sync"
)

// ConcurrentList — потокобезопасная, строго типизированная (Generic) реализация динамического списка.
//
// Спроектирована для сценариев с высокой интенсивностью конкурентного чтения и пакетной обработки.
// Интегрирована с нативным движком итераторов Go (пакет "iter"), что позволяет прозрачно
// использовать стандартные циклы `for range` без риска вызвать состояние гонки (Race Condition).
type ConcurrentList[T any] struct {
	mu    sync.RWMutex // RWMutex оптимизирует параллельное чтение множеством горутин
	items []T          // Внутреннее несинхронизированное хранилище элементов
}

// NewConcurrentList — фабричный конструктор конкурентного списка.
func NewConcurrentList[T any]() *ConcurrentList[T] {
	return &ConcurrentList[T]{
		items: make([]T, 0),
	}
}

// Append атомарно добавляет элемент в конец списка под эксклюзивной write-блокировкой.
func (l *ConcurrentList[T]) Append(item T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, item)
}

// Get извлекает элемент по его индексу под read-блокировкой.
// Возвращает zero-value типа T и флаг false, если индекс вышел за фактические границы слайса.
func (l *ConcurrentList[T]) Get(index int) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var zero T
	if index < 0 || index >= len(l.items) {
		return zero, false
	}
	return l.items[index], true
}

// Len возвращает текущее атомарное количество элементов в списке.
func (l *ConcurrentList[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.items)
}

// Snapshot генерирует изолированную, независимую копию текущего состояния списка.
// Рекомендуется для тяжелых фоновых операций маршалинга, сортировки или выгрузки дампов.
func (l *ConcurrentList[T]) Snapshot() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	cp := make([]T, len(l.items))
	copy(cp, l.items)
	return cp
}

// All возвращает нативный функциональный итератор iter.Seq2, отдающий пары [индекс, значение].
//
// Позволяет писать идиоматичные циклы:
//
//	for idx, val := range list.All() { ... }
//
// Безопасность рантайма: метод делает мгновенный снимок под RLock, после чего отпускает мьютекс.
// Бизнес-логика внутри цикла работает с копией, не блокируя горутины-писатели на время итерации.
func (l *ConcurrentList[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		// 1. Делаем быстрый снимок данных под RLock, чтобы минимизировать время удержания блокировки.
		l.mu.RLock()
		snapshot := make([]T, len(l.items))
		copy(snapshot, l.items)
		l.mu.RUnlock()

		// 2. Итерируемся по изолированной копии.
		for idx, val := range snapshot {
			// yield возвращает false, если пользователь вызвал break внутри цикла for range.
			// В этом случае мгновенно прерываем выполнение замыкания.
			if !yield(idx, val) {
				return
			}
		}
	}
}

// Values возвращает нативный функциональный итератор iter.Seq, отдающий только значения.
//
// Позволяет писать циклы:
//
//	for val := range list.Values() { ... }
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
