package cache

import (
	"container/list"
	"sync"
)

// LRUEvict реализует интерфейс EvictionPolicy[K], используя стратегию Least Recently Used (Наименее недавно использовавшийся).
//
// Вытесняет элементы, к которым дольше всего не было обращений (как на чтение, так и на запись).
// На Highload-нагрузках обеспечивает O(1) производительность за счет связки двусвязного списка хронологии запросов
// и индексированной хэш-мапы быстрых указателей на элементы списка.
type LRUEvict[K comparable] struct {
	mu       sync.Mutex          // Исключительный Mutex защищает внутренние структуры от Race Condition
	ll       *list.List          // Очередь свежести использования: самые "горячие" в Front(), "холодные" в Back()
	items    map[K]*list.Element // Хэш-карта быстрых указателей на узлы списка для мгновенного доступа
	emptyKey K                   // Заглушка zero-value типа K для безопасных возвратов при пустом кэше
}

// NewLRUEvict — фабричный构造тель стратегии вытеснения LRU.
func NewLRUEvict[K comparable]() *LRUEvict[K] {
	return &LRUEvict[K]{
		ll:    list.New(),
		items: make(map[K]*list.Element),
	}
}

// OnGet регистрирует факт успешного чтения элемента.
// Находит запись и перемещает её в голову очереди (Front), продлевая ей время жизни в кэше.
func (lru *LRUEvict[K]) OnGet(key K) {
	lru.mu.Lock() // Только Lock, только победа!
	defer lru.mu.Unlock()
	if el, ok := lru.items[key]; ok {
		lru.ll.MoveToFront(el)
	}
}

// OnSet фиксирует событие вставки или модификации записи.
// Если ключ уже есть — перемещает в голову (омолаживает). Если новый — пушит в Front().
func (lru *LRUEvict[K]) OnSet(key K) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// Если ключ уже присутствует в памяти — обновляем его позицию в очереди свежести
	if el, ok := lru.items[key]; ok {
		lru.ll.MoveToFront(el)
		return
	}

	// Если ключ абсолютно новый — фиксируем его в голове списка
	lru.items[key] = lru.ll.PushFront(key)
}

// OnRemove принудительно стирает ключ из очереди и хэш-индексов, предотвращая утечки памяти.
func (lru *LRUEvict[K]) OnRemove(key K) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	lru.remove(key) // Безопасный вызов приватного неблокирующего метода
}

// Reset полностью очищает внутреннее состояние контроллера LRU-памяти.
func (lru *LRUEvict[K]) Reset() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.ll.Init()
	lru.items = make(map[K]*list.Element)
}

// Evict вычисляет и атомарно вырезает из памяти самую «холодную» запись, находящуюся в хвосте очереди (Back).
// Возвращает zero-value и флаг false, если хранилище пустое.
func (lru *LRUEvict[K]) Evict() (K, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	el := lru.ll.Back()
	if el == nil {
		return lru.emptyKey, false
	}

	// Распаковываем строго типизированный ключ из интерфейсного значения Value
	key := el.Value.(K)
	lru.remove(key) // Выполняется в рамках текущего лока — дедлок полностью исключен!
	return key, true
}

// ====================================================================
// Внутренние неблокирующие методы (Helper methods — выполняются строго под локом)
// ====================================================================

// remove осуществляет непосредственное вырезание узла из двусвязного списка и хэш-карты.
func (lru *LRUEvict[K]) remove(key K) {
	if el, ok := lru.items[key]; ok {
		lru.ll.Remove(el)
		delete(lru.items, key)
	}
}
