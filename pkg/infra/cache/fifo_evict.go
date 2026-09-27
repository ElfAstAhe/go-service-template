package cache

import (
	"container/list"
	"sync"
)

// FIFOEvict реализует интерфейс EvictionPolicy[K], используя стратегию First-In, First-Out (Первым пришел — первым ушел).
//
// Вытесняет элементы строго в порядке их добавления в кэш-хранилище, независимо от частоты или свежести их чтения.
// Для обеспечения O(1) производительности комбинирует двусвязный список (очередь) и индексированную хэш-мапу.
type FIFOEvict[K comparable] struct {
	mu    sync.RWMutex        // RWMutex защищает внутренние структуры от Race Condition
	ll    *list.List          // Очередь хронологии добавления элементов
	items map[K]*list.Element // Хэш-карта быстрых указателей на элементы списка для O(1) доступа
}

// NewFIFOEvict — фабричный конструктор стратегии вытеснения FIFO.
func NewFIFOEvict[K comparable]() *FIFOEvict[K] {
	return &FIFOEvict[K]{
		ll:    list.New(),
		items: make(map[K]*list.Element),
	}
}

// OnGet регистрирует факт чтения элемента. В классическом FIFO операция чтения
// полностью игнорируется и никак не влияет на приоритет вытеснения записи.
func (fif *FIFOEvict[K]) OnGet(key K) {}

// OnSet фиксирует событие вставки нового ключа.
// Элемент атомарно проталкивается в голову очереди (Front), фиксируя свое время рождения.
func (fif *FIFOEvict[K]) OnSet(key K) {
	fif.mu.Lock()
	defer fif.mu.Unlock()

	// Если ключ уже присутствует, в классической схеме FIFO его позиция и возраст не обновляются
	if _, ok := fif.items[key]; !ok {
		fif.items[key] = fif.ll.PushFront(key)
	}
}

// OnRemove принудительно вычищает указанный ключ из очереди и хэш-индексов стратегии,
// полностью защищая рантайм от скрытых утечек памяти (Memory Leaks).
func (fif *FIFOEvict[K]) OnRemove(key K) {
	fif.mu.Lock()
	defer fif.mu.Unlock()

	if el, ok := fif.items[key]; ok {
		fif.ll.Remove(el)
		delete(fif.items, key)
	}
}

// Reset обнуляет состояние контроллера вытеснения, эффективно переиспользуя аллоцированную под мапу память.
func (fif *FIFOEvict[K]) Reset() {
	fif.mu.Lock()
	defer fif.mu.Unlock()

	fif.ll.Init()
	clear(fif.items) // Go 1.21+ оптимизация: сброс мапы без деаллокации её бакетов
}

// Evict вычисляет и возвращает самую «старую» запись в кэше, находящуюся в хвосте очереди (Back).
// Возвращает zero-value и флаг false, если хранилище пустое.
func (fif *FIFOEvict[K]) Evict() (K, bool) {
	// Самый возрастной элемент всегда находится в самом хвосте, так как новые мы пушим в голову (Front)
	el := fif.ll.Back()
	if el == nil {
		var zero K
		return zero, false
	}

	// Извлекаем строго типизированный ключ из интерфейсного значения Value
	key := el.Value.(K)
	fif.OnRemove(key) // Очищаем внутренние индексы стратегии под защитой мьютекса

	return key, true
}
