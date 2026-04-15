package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
)

// Бенчмарк для сравнения Lock Contention (Борьба за мьютекс)
func BenchmarkStorage_Contention(b *testing.B) {
	data := []byte("some-heavy-payload")

	// Сценарий 1: Один мьютекс (RawStorage)
	b.Run("RawStorage_SingleMutex", func(b *testing.B) {
		s := cache.NewRawStorage[string](10000, cache.NewLRUEvict[string]())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i)
				s.Set(key, data)
				_, _ = s.Get(key)
				i++
			}
		})
	})

	// Сценарий 2: Шардирование (ShardStorage)
	b.Run("ShardStorage_64Shards", func(b *testing.B) {
		factory := func(m int, p cache.EvictionPolicy[string]) cache.Storage[string] {
			return cache.NewRawStorage[string](m, p)
		}
		// Используем 64 шарда (степень двойки для скорости)
		s := cache.NewShardStorage[string](64, factory, 10000, cache.NewLRUEvict[string]())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i)
				s.Set(key, data)
				_, _ = s.Get(key)
				i++
			}
		})
	})
}

// Бенчмарк полного цикла Manager
func BenchmarkManager_FullCycle(b *testing.B) {
	factory := func() BenchData { return BenchData{} }
	val := BenchData{ID: 1, Value: "benchmark-test-payload"}

	// JSON vs GOB
	codecs := []struct {
		name  string
		codec cache.Codec[BenchData]
	}{
		{"JSON", cache.NewJSONCodec(factory)},
		{"GOB", cache.NewGobCodec(factory)},
	}

	for _, tc := range codecs {
		b.Run(tc.name, func(b *testing.B) {
			c, _ := cache.CacheFactory(
				cache.WithShardCount[string, BenchData](64),
				cache.WithCodec[string, BenchData](tc.codec),
				cache.WithLRUEvictPolicy[string, BenchData](),
			)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("key-%d", i)
					_ = c.Set(key, val, time.Minute)
					_, _, _ = c.Get(key)
					i++
				}
			})
		})
	}
}
