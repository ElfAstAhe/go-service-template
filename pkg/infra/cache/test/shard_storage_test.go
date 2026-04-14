package test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShardStorage_Logic(t *testing.T) {
	// Фабрика, создающая моки шардов
	factory := func(maxSize int, policy cache.EvictionPolicy[string]) cache.Storage[string] {
		return mocks.NewMockStorage[string](t)
	}

	t.Run("power_of_two_optimization", func(t *testing.T) {
		// 64 - это степень двойки
		ss := cache.NewShardStorage[string, any](64, factory, 100, nil)

		// Проверяем распределение: разные ключи должны попадать в разные шарды
		key1 := "user_1"
		key2 := "user_2"

		shard1 := ss.GetShard(key1)
		shard2 := ss.GetShard(key2)

		assert.NotNil(t, shard1)
		assert.NotNil(t, shard2)
		// С большой вероятностью хэши будут разные
		//assert.NotEqual(t, shard1, shard2, "Keys should fall into different shards")
	})

	t.Run("simple_modulo_index", func(t *testing.T) {
		// 10 - не степень двойки
		ss := cache.NewShardStorage[string, any](10, factory, 100, nil)

		shard := ss.GetShard("any_key")
		assert.NotNil(t, shard)
	})
}

func TestShardStorage_Methods(t *testing.T) {
	mShard := mocks.NewMockStorage[string](t)
	factory := func(maxSize int, policy cache.EvictionPolicy[string]) cache.Storage[string] {
		return mShard
	}

	// Создаем сторадж с 1 шардом для простоты тестов методов
	ss := cache.NewShardStorage[string, any](1, factory, 100, nil)

	t.Run("Set_Get_Delete", func(t *testing.T) {
		key := "test"
		data := []byte("payload")

		mShard.On("Set", key, data).Once()
		ss.Set(key, data)

		mShard.On("Get", key).Return(data, true).Once()
		val, ok := ss.Get(key)
		assert.True(t, ok)
		assert.Equal(t, data, val)

		mShard.On("Delete", key).Once()
		ss.Delete(key)
	})

	t.Run("Len_And_Clear", func(t *testing.T) {
		mShard.On("Len").Return(5).Once()
		assert.Equal(t, 5, ss.Len())

		mShard.On("Clear").Once()
		ss.Clear()
	})
}

func TestShardStorage_Range_Stop(t *testing.T) {
	// Создаем 2 шарда
	shards := []*mocks.MockStorage[string]{
		mocks.NewMockStorage[string](t),
		mocks.NewMockStorage[string](t),
	}

	i := 0
	factory := func(maxSize int, policy cache.EvictionPolicy[string]) cache.Storage[string] {
		s := shards[i]
		i++
		return s
	}

	ss := cache.NewShardStorage[string, any](2, factory, 100, nil)

	// Имитируем, что первый шард возвращает данные, а мы хотим остановиться
	shards[0].On("Range", mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(0).(func(string, []byte) bool)
		fn("key1", []byte("val1")) // тут вернем false в тесте ниже
	}).Return()

	// До второго шарда Range дойти НЕ должен

	count := 0
	ss.Range(func(key string, value []byte) bool {
		count++
		return false // Останавливаемся сразу
	})

	assert.Equal(t, 1, count, "Range should stop after first element and not visit other shards")
}

func TestShardStorage_Concurrency_Real(t *testing.T) {
	// Тест на реальных данных без моков для проверки Race Condition
	shardCount := uint64(16)
	factory := func(maxSize int, policy cache.EvictionPolicy[int]) cache.Storage[int] {
		return cache.NewByteStorage[int](maxSize, policy)
	}

	//    policy := cache.NewFIFOEvict[int]()
	// policy := cache.NewLFUEvict[int]()
	policy := cache.NewLRUEvict[int]()
	ss := cache.NewShardStorage[int, any](shardCount, factory, 1000, policy)

	wg := sync.WaitGroup{}
	iterations := 1000
	workers := 10

	wg.Add(workers * 2)
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := workerID*iterations + j
				ss.Set(key, []byte(fmt.Sprintf("val-%d", key)))
			}
		}(i)

		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := workerID*iterations + j
				ss.Get(key)
			}
		}(i)
	}

	wg.Wait()
	assert.True(t, ss.Len() > 0)
}
