package utils

import (
	"fmt"
	"testing"
)

// Нагрузочный бенчмарк: симулирует параллельное чтение через All() и запись через Append
func BenchmarkConcurrentList_ParallelReadWrite(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("SliceSize-%d", size), func(b *testing.B) {
			list := NewConcurrentList[int]()
			for i := 0; i < size; i++ {
				list.Append(i)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				// Каждая горутина либо читает, либо пишет в соотношении 90% на 10%
				i := 0
				for pb.Next() {
					if i%10 == 0 {
						list.Append(i)
					} else {
						// Тестируем производительность итератора All()
						for _, val := range list.All() {
							_ = val
						}
					}
					i++
				}
			})
		})
	}
}

// Бенчмарк для сравнения накладных расходов: Snapshot() против All()
func BenchmarkConcurrentList_SnapshotVsIterator(b *testing.B) {
	size := 500
	list := NewConcurrentList[int]()
	for i := 0; i < size; i++ {
		list.Append(i)
	}

	b.Run("Snapshot", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			snapshot := list.Snapshot()
			for _, val := range snapshot {
				_ = val
			}
		}
	})

	b.Run("Iterator-All", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, val := range list.All() {
				_ = val
			}
		}
	})
}
