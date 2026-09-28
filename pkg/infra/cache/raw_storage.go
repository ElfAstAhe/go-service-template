package cache

import (
	"sync"
)

// RawStorage — базовая потокобезопасная реализация несегментированного кэш-хранилища (In-Memory Storage).
//
// Инкапсулирует в себе сырую хэш-карту Go, защищенную мьютексом (sync.RWMutex), и координирует
// вызовы хуков выбранного контроллера вытеснения данных (EvictionPolicy) при каждой CRUD-операции.
type RawStorage[K comparable] struct {
	mu      sync.RWMutex      // Мьютекс для защиты конкурентного доступа со стороны множества горутин
	data    map[K][]byte      // Прямое несинхронизированное хранилище бинарных payload-пакетов
	policy  EvictionPolicy[K] // Инжектированная математическая стратегия вытеснения ключей (LRU/LFU/FIFO)
	maxSize int               // Жесткий потолок емкости (DataCapacity) для предотвращения переполнения RAM
}

// NewRawStorage создает новый экземпляр базового кэш-хранилища
func NewRawStorage[K comparable](maxSize int, policy EvictionPolicy[K]) *RawStorage[K] {
	return &RawStorage[K]{
		data:    make(map[K][]byte),
		maxSize: maxSize,
		policy:  policy,
	}
}

// Get извлекает бинарный payload по его ключу.
// ⚠️ Важный рантайм-нюанс: метод использует эксклюзивный Lock() вместо RLock(), так как хук OnGet
// внутри стратегий LRU/LFU выполняет модификацию памяти (перестановку или инкремент элементов).
func (rs *RawStorage[K]) Get(key K) ([]byte, bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	b, ok := rs.data[key]
	if ok {
		rs.policy.OnGet(key) // Омолаживаем или инкрементируем частоту ключа в стратегии вытеснения
	}

	return b, ok
}

// Set сохраняет бинарный payload в память.
// Если лимит maxSize превышен, метод атомарно вычисляет и вырезает из памяти «жертву» (Evict).
func (rs *RawStorage[K]) Set(key K, b []byte) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Если ключа нет, проверяем, не пора ли кого-то выселить
	if _, exists := rs.data[key]; !exists {
		if rs.maxSize > 0 && len(rs.data) >= rs.maxSize {
			if victim, ok := rs.policy.Evict(); ok {
				delete(rs.data, victim) // Удаляем вытесненный элемент из физической мапы
			}
		}
	}

	rs.data[key] = b
	rs.policy.OnSet(key) // Фиксируем факт вставки/обновления в индексах стратегии
}

// Delete транзакционно удаляет запись из мапы и очищает связанные индексы контроллера вытеснения.
func (rs *RawStorage[K]) Delete(key K) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	delete(rs.data, key)
	rs.policy.OnRemove(key) // Защита от скрытых утечек памяти (Memory Leaks) внутри стратегии
}

// Range выполняет потоковый обход (итерацию) всех элементов под защитой read-блокировки.
// Блокирует операции записи (Set/Delete), но разрешает параллельные вызовы атомарных проверок (Has/Len).
func (rs *RawStorage[K]) Range(fn func(key K, value []byte) bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	for k, v := range rs.data {
		if !fn(k, v) {
			break
		}
	}
}

// Has проверяет наличие ключа под быстрой RLock-блокировкой без влияния на приоритеты вытеснения.
func (rs *RawStorage[K]) Has(key K) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, ok := rs.data[key]
	return ok
}

// Len возвращает текущее атомарное количество элементов, удерживаемых в текущем сегменте памяти.
func (rs *RawStorage[K]) Len() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.data)
}

// Clear выполняет полный сброс памяти хранилища и обнуляет внутренние структуры контроллера вытеснения.
func (rs *RawStorage[K]) Clear() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.data = make(map[K][]byte)
	rs.policy.Reset() // Сбрасываем очереди и списки частот
}
