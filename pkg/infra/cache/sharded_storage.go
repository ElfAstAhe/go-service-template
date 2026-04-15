package cache

import (
	"fmt"
	"hash/maphash"
	"sync"
	"unsafe"
)

type ShardFactory[K comparable] func(maxSize int, policy EvictionPolicy[K]) Storage[K]
type ShardIndex[K comparable] func(K) uint64

type ShardStorage[K comparable] struct {
	hashSeed   maphash.Seed
	hashPool   *sync.Pool
	shards     []Storage[K]
	shardIndex ShardIndex[K]
	shardCount uint64
}

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
	res.shardIndex = res.ShardIndexSelector(res.shardCount)

	for i := uint64(0); i < res.shardCount; i++ {
		res.shards = append(res.shards, shardFactory(maxSize, policy))
	}

	return res
}

func (ss *ShardStorage[K]) Get(key K) ([]byte, bool) {
	return ss.GetShard(key).Get(key)
}

func (ss *ShardStorage[K]) Set(key K, b []byte) {
	ss.GetShard(key).Set(key, b)
}

func (ss *ShardStorage[K]) Delete(key K) {
	ss.GetShard(key).Delete(key)
}

func (ss *ShardStorage[K]) Has(key K) bool {
	return ss.GetShard(key).Has(key)
}

func (ss *ShardStorage[K]) Len() int {
	var total int
	// Собираем длину со всех шардов.
	// Так как каждый шард внутри под своим мьютексом, это безопасно.
	for _, shard := range ss.shards {
		total += shard.Len()
	}
	return total
}

func (ss *ShardStorage[K]) Clear() {
	// Очищаем каждый шард по очереди
	for _, shard := range ss.shards {
		shard.Clear()
	}
}

func (ss *ShardStorage[K]) Range(fn func(key K, value []byte) bool) {
	// Последовательно итерируем каждый шард.
	// Если пользовательская функция fn вернет false,
	// мы полностью прерываем обход всех шардов.
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

// ShardIndexSelector выбор стратегии расчёта индекса шарда
func (ss *ShardStorage[K]) ShardIndexSelector(shardCount uint64) ShardIndex[K] {
	if ss.isPowerOfTwo(shardCount) {
		return ss.powerOfTwoShardIndex
	}

	return ss.simpleShardIndex
}

func (ss *ShardStorage[K]) isPowerOfTwo(n uint64) bool {
	return n > 0 && (n&(n-1)) == 0
}

func (ss *ShardStorage[K]) GetShard(key K) Storage[K] {
	return ss.shards[ss.shardIndex(key)]
}

func (ss *ShardStorage[K]) keyHasher(key K) uint64 {
	h := ss.hashPool.Get().(*maphash.Hash)
	defer ss.hashPool.Put(h)

	h.Reset()
	h.SetSeed(ss.hashSeed)

	// Оптимизированный путь для базовых типов
	switch v := any(key).(type) {
	case string:
		_, _ = h.WriteString(v)
	case int, uint, int64, uint64, int32, uint32, float64:
		// Создаем слайс байтов прямо из области памяти переменной
		size := unsafe.Sizeof(v)
		b := unsafe.Slice((*byte)(unsafe.Pointer(&v)), size)
		_, _ = h.Write(b)
	default:
		// Для сложных типов, где есть указатели (например, слайсы или мапы внутри),
		// unsafe.Slice брать нельзя — он хеширует только заголовки (адреса).
		// Поэтому тут оставляем надежный fmt.
		// Самый медленный способ записи массива байт
		_, _ = fmt.Fprint(h, v)
	}

	return h.Sum64()
}

// simpleShardIndex простой:  остаток от деления
func (ss *ShardStorage[K]) simpleShardIndex(key K) uint64 {
	return ss.keyHasher(key) % ss.shardCount
}

// powerOfTwoShardIndex кратно степени 2, если shardCount = 64, то 64-1 = 63 (в битах это 111111)
// Побитовое И мгновенно дает индекс
func (ss *ShardStorage[K]) powerOfTwoShardIndex(key K) uint64 {
	return ss.keyHasher(key) & (ss.shardCount - 1)
}
