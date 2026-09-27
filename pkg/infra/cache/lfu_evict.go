package cache

import (
	"container/list"
	"sync"
)

// LFUItem инкапсулирует полезную нагрузку элемента кэша и счетчик интенсивности обращений к нему.
type LFUItem[K comparable] struct {
	key   K   // Уникальный строго типизированный ключ элемента
	count int // Счетчик (частота) вызовов операций чтения/записи для данного ключа
}

// LFUEvict реализует интерфейс EvictionPolicy[K], используя алгоритм Least Frequently Used (Наименее часто используемый).
//
// 💡 Алгоритмический нюанс:
// Для достижения экстремальной производительности O(1) на всех операциях, структура комбинирует
// индексированную хэш-карту быстрых указателей и мапу независимых двусвязных списков, сгруппированных
// по частоте обращений (freqs). Это исключает необходимость сортировки памяти на Highload-нагрузках.
type LFUEvict[K comparable] struct {
	mu       sync.Mutex          // Исключительный Mutex защищает структуры от Race Condition в параллельных горутинах
	items    map[K]*list.Element // Карта быстрого O(1) доступа к элементам в списках частот
	freqs    map[int]*list.List  // Матрица списков: Ключ — частота (hits), Значение — двусвязный список элементов
	minFreq  int                 // Глобальный указатель на текущую минимальную частоту для моментального вытеснения
	emptyKey K                   // Заглушка zero-value типа K для безопасных возвратов при пустом кэше
}

// NewLFUEvict — фабричный конструктор стратегии вытеснения LFU.
func NewLFUEvict[K comparable]() *LFUEvict[K] {
	return &LFUEvict[K]{
		items: make(map[K]*list.Element),
		freqs: make(map[int]*list.List),
	}
}

// OnGet регистрирует факт успешного чтения элемента.
// Находит запись и атомарно продвигает её счетчик частоты вверх с ребалансировкой списков.
func (lfu *LFUEvict[K]) OnGet(key K) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	el, ok := lfu.items[key]
	if !ok {
		return
	}

	lfu.increment(el)
}

// OnSet фиксирует событие создания или обновления записи.
// Если ключ уже存在 — инкрементирует частоту. Если новый — инициализирует его с частотой 1.
func (lfu *LFUEvict[K]) OnSet(key K) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if el, ok := lfu.items[key]; ok {
		lfu.increment(el)
		return
	}

	// Инициализация новой сущности: частота старта всегда равна 1
	item := &LFUItem[K]{key: key, count: 1}
	lfu.minFreq = 1 // Сбрасываем глобальный указатель минимума на единицу
	lfu.insert(item)
}

// OnRemove принудительно стирает ключ из всех списков частот и мап, предотвращая утечки памяти.
func (lfu *LFUEvict[K]) OnRemove(key K) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.remove(key)
}

// Evict вычисляет и атомарно вырезает из памяти наименее востребованный элемент (жертву).
// Выборка происходит из хвоста (Back) списка с минимальной частотой minFreq за константное время O(1).
func (lfu *LFUEvict[K]) Evict() (K, bool) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if len(lfu.items) == 0 {
		return lfu.emptyKey, false
	}

	// Извлекаем список элементов с наименьшей частотой запросов
	lst := lfu.freqs[lfu.minFreq]
	el := lst.Back() // Самый старый элемент внутри данной частоты лежит в самом хвосте
	if el == nil {
		return lfu.emptyKey, false
	}

	key := el.Value.(*LFUItem[K]).key
	lfu.remove(key) // Вызываем внутренний неблокирующий метод удаления ресурсов

	return key, true
}

// Reset полностью очищает внутреннее состояние контроллера LFU-памяти.
func (lfu *LFUEvict[K]) Reset() {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.items = make(map[K]*list.Element)
	lfu.freqs = make(map[int]*list.List)
	lfu.minFreq = 0
}

// ====================================================================
// Внутренние неблокирующие методы (Helper methods — выполняются строго под локом)
// ====================================================================

// increment переводит элемент на следующий уровень частоты.
func (lfu *LFUEvict[K]) increment(el *list.Element) {
	item := el.Value.(*LFUItem[K])
	oldFreq := item.count

	// Удаляем элемент из текущего (старого) списка частоты
	lfu.freqs[oldFreq].Remove(el)

	// Если старый список опустел и текущий минимум указывал на него — смещаем глобальный минимум вверх
	if lfu.freqs[oldFreq].Len() == 0 && oldFreq == lfu.minFreq {
		lfu.minFreq++
	}

	item.count++
	lfu.insert(item) // Переносим объект в новый список частоты
}

// insert подкладывает объект в голову (Front) списка соответствующей частоты.
func (lfu *LFUEvict[K]) insert(item *LFUItem[K]) {
	if _, ok := lfu.freqs[item.count]; !ok {
		lfu.freqs[item.count] = list.New()
	}
	// Пушим в голову: внутри одной частоты самыми свежими будут элементы в начале списка
	el := lfu.freqs[item.count].PushFront(item)
	lfu.items[item.key] = el
}

// remove осуществляет непосредственное стирание ключа из списков и хэш-карт.
func (lfu *LFUEvict[K]) remove(key K) {
	if el, ok := lfu.items[key]; ok {
		item := el.Value.(*LFUItem[K])
		lfu.freqs[item.count].Remove(el)
		delete(lfu.items, key)
	}
}
