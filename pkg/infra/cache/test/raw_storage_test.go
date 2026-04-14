package test

import (
	"sync"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	"github.com/stretchr/testify/assert"
)

//func TestRawStorage_EvictionTrigger(t *testing.T) {
//    // Проверяем, что при достижении maxSize вызывается Evict у политики
//    mPolicy := mocks.NewMockEvictionPolicy[string](t)
//    storage := cache.NewByteStorage[string](2, mPolicy)
//
//    // Первый Set
//    mPolicy.On("OnSet", "k1").Return().Once()
//    storage.Set("k1", []byte("v1"))
//
//    // Второй Set
//    mPolicy.On("OnSet", "k2").Return().Once()
//    storage.Set("k2", []byte("v2"))
//
//    // Третий Set — должен сработать лимит
//    // Сначала Storage спрашивает у политики, кого выкинуть
//    mPolicy.On("Evict").Return("k1", true).Once()
//    // Затем уведомляет, что ключ k1 удален
//    mPolicy.On("OnRemove", "k1").Return().Once()
//    // И только потом добавляет новый
//    mPolicy.On("OnSet", "k3").Return().Once()
//
//    storage.Set("k3", []byte("v3"))
//
//    assert.Equal(t, 2, storage.Len())
//    assert.False(t, storage.Has("k1"))
//    assert.True(t, storage.Has("k3"))
//}

func TestRawStorage_Get_TriggersPolicy(t *testing.T) {
	mPolicy := mocks.NewMockEvictionPolicy[string](t)
	storage := cache.NewByteStorage[string](10, mPolicy)

	mPolicy.On("OnSet", "key").Return()
	storage.Set("key", []byte("val"))

	// Проверяем, что Get дергает OnGet (важно для LRU/LFU)
	mPolicy.On("OnGet", "key").Return().Once()

	_, ok := storage.Get("key")
	assert.True(t, ok)
}

func TestRawStorage_Concurrency(t *testing.T) {
	// Самый важный тест — на отсутствие race condition
	// Используем реальную политику (например, FIFO), чтобы не мучиться с моками в гонках
	policy := cache.NewFIFOEvict[int]()
	storage := cache.NewByteStorage[int](100, policy)

	const goroutines = 50
	const opsPerGoro = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // Set, Get и Delete

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoro; j++ {
				storage.Set(id*opsPerGoro+j, []byte("data"))
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoro; j++ {
				storage.Get(id*opsPerGoro + j)
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoro; j++ {
				storage.Delete(id*opsPerGoro + j)
			}
		}(i)
	}

	wg.Wait()
	// Если мы здесь и не поймали panic/race — тест пройден
}

func TestRawStorage_Range_Break(t *testing.T) {
	policy := cache.NewFIFOEvict[string]()
	storage := cache.NewByteStorage[string](10, policy)

	storage.Set("a", []byte("1"))
	storage.Set("b", []byte("2"))
	storage.Set("c", []byte("3"))

	counter := 0
	storage.Range(func(key string, value []byte) bool {
		counter++
		return false // Прерываем после первого элемента
	})

	assert.Equal(t, 1, counter)
}
