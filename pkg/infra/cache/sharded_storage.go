package cache

import (
	"fmt"
	"hash/maphash"
	"sync"
	"unsafe"
)

// ShardFactory определяет сигнатуру функции-фабрики для создания единичного изолированного сегмента (шарда) кэша.
type ShardFactory[K comparable] func(maxSize int, policy EvictionPolicy[K]) Storage[K]

// ShardIndex определяет сигнатуру математической функции распределения ключей по индексам шардов.
type ShardIndex[K comparable] func(K) uint64

// ShardStorage реализует интерфейс Storage[K], представляя собой горизонтально масштабируемое,
// сегментированное (шардированное) ин-мемори хранилище данных кэша.
//
// 💡 Архитектурный паттерн (Striped Locking / Sharding):
// Разрезает единую карту памяти на пул независимых сегментов (shards), каждый из которых управляется
// своим локальным мьютексом. Это кратно снижает конкуренцию за блокировки (Mutex Contention)
// на Highload-нагрузках, позволяя параллельным горутинам читать и писать в кэш без деградации RPS.
type ShardStorage[K comparable] struct {
	hashSeed   maphash.Seed  // Уникальная рандомная соль хэширования, генерируемая при старте для защиты от атак Hash Flooding
	hashPool   *sync.Pool    // Пул потокобезопасных объектов maphash.Hash для обеспечения Zero-Allocation вычислений хэшей
	shards     []Storage[K]  // Слайс изолированных внутренних сегментов кэша (обычно RawStorage)
	shardIndex ShardIndex[K] // Активная рантайм-стратегия вычисления целевого индекса шарда
	shardCount uint64        // Общее количество сконфигурированных шардов
}

// NewShardStorage — фабричный конструктор шардированного хранилища кэша.
func NewShardStorage[K comparable](
	shardCount uint64,
	shardFactory ShardFactory[K],
	maxSize int,
	policy EvictionPolicy[K],
) *ShardStorage[K] {
	res := &ShardStorage[K]{
		hashSeed: maphash.MakeSeed(),
		hashPool: &sync.Pool{
			New: func() any {
				return new(maphash.Hash)
			},
		},
		shardCount: shardCount,
		shards:     make([]Storage[K], 0, shardCount),
	}
	// Динамически выбираем наиболее производительный алгоритм расчета индекса
	res.shardIndex = res.ShardIndexSelector(res.shardCount)

	// Инициализируем изолированные сегменты памяти
	for i := uint64(0); i < res.shardCount; i++ {
		res.shards = append(res.shards, shardFactory(maxSize, policy))
	}

	return res
}

// Get вычисляет целевой шард и извлекает из него бинарный payload.
func (ss *ShardStorage[K]) Get(key K) ([]byte, bool) {
	return ss.GetShard(key).Get(key)
}

// Set вычисляет целевой шард и атомарно сохраняет в него бинарный payload.
func (ss *ShardStorage[K]) Set(key K, b []byte) {
	ss.GetShard(key).Set(key, b)
}

// Delete вычисляет целевой шард и принудительно стирает ключ с очисткой его индексов вытеснения.
func (ss *ShardStorage[K]) Delete(key K) {
	ss.GetShard(key).Delete(key)
}

// Has проверяет наличие ключа в целевом сегменте памяти без изменения приоритетов вытеснения.
func (ss *ShardStorage[K]) Has(key K) bool {
	return ss.GetShard(key).Has(key)
}

// Len собирает текущий размер данных со всех шардов. Отрабатывает конкурентно-безопасно,
// так как каждый сегмент опрашивается под своим собственным внутренним локальным мьютексом.
func (ss *ShardStorage[K]) Len() int {
	var total int
	for _, shard := range ss.shards {
		total += shard.Len()
	}
	return total
}

// Clear последовательно сбрасывает и обнуляет память каждого шарда по отдельности.
func (ss *ShardStorage[K]) Clear() {
	for _, shard := range ss.shards {
		shard.Clear()
	}
}

// Range выполняет сквозной последовательный обход всех шардов кэша.
// Если переданная функция-колбек fn вернет false, глобальный цикл обхода мгновенно прерывается.
func (ss *ShardStorage[K]) Range(fn func(key K, value []byte) bool) {
	for _, shard := range ss.shards {
		stop := false
		shard.Range(func(key K, value []byte) bool {
			if !fn(key, value) {
				stop = true
				return false
			}
			return true
		})

		if stop {
			break
		}
	}
}

// ShardIndexSelector анализирует архитектуру емкости: если shardCount кратен степени двойки,
// подключает ультра-скоростную стратегию битовых масок, иначе откатывается на стандартный остаток от деления.
func (ss *ShardStorage[K]) ShardIndexSelector(shardCount uint64) ShardIndex[K] {
	if ss.isPowerOfTwo(shardCount) {
		return ss.powerOfTwoShardIndex
	}

	return ss.simpleShardIndex
}

// isPowerOfTwo проверяет, является ли число степенью двойки с помощью побитовой маски.
func (ss *ShardStorage[K]) isPowerOfTwo(n uint64) bool {
	return n > 0 && (n&(n-1)) == 0
}

// GetShard возвращает ссылку на конкретный изолированный сегмент Storage, обслуживающий данный ключ.
func (ss *ShardStorage[K]) GetShard(key K) Storage[K] {
	return ss.shards[ss.shardIndex(key)]
}

// keyHasher вычисляет 64-битный некриптографический хэш от ключа произвольного типа.
// Использует оптимизации unsafe-копирования памяти для примитивов для полного исключения аллокаций.
func (ss *ShardStorage[K]) keyHasher(key K) uint64 {
	h := ss.hashPool.Get().(*maphash.Hash)
	defer ss.hashPool.Put(h)

	h.Reset()
	h.SetSeed(ss.hashSeed)

	switch v := any(key).(type) {
	case string:
		_, _ = h.WriteString(v)
	case int, uint, int64, uint64, int32, uint32, float64:
		// Высокопроизводительный Fast-Path: отображаем область памяти числовой переменной
		// напрямую в слайс байт, полностью минуя кучу (Zero-Heap Allocation)
		size := unsafe.Sizeof(v)
		//nolint:gosec // G103:
		b := unsafe.Slice((*byte)(unsafe.Pointer(&v)), size)
		_, _ = h.Write(b)
	default:
		// Slow-Path: для сложных структур со вложенными указателями используем надежный fmt.Fprint
		_, _ = fmt.Fprint(h, v)
	}

	return h.Sum64()
}

// simpleShardIndex вычисляет индекс сегмента по классической формуле остатка от деления (Modulo).
func (ss *ShardStorage[K]) simpleShardIndex(key K) uint64 {
	return ss.keyHasher(key) % ss.shardCount
}

// powerOfTwoShardIndex вычисляет индекс сегмента с помощью побитового И (Bitwise AND).
// Экстремально эффективная инструкция CPU, заменяющая дорогое деление при емкостях, кратных степени 2 (например, 64 шардов).
func (ss *ShardStorage[K]) powerOfTwoShardIndex(key K) uint64 {
	return ss.keyHasher(key) & (ss.shardCount - 1)
}
