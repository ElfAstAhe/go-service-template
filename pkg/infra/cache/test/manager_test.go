package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Тест успешного получения данных
func TestManager_Get_Success(t *testing.T) {
	mStorage := mocks.NewMockStorage[string](t)
	mCodec := mocks.NewMockCodec[string](t)
	mgr := cache.New[string, string](mStorage, mCodec, 0)

	key := "test-key-1"
	expectedVal := "test-value-data"
	raw := []byte("raw-bytes")

	mStorage.On("Get", key).Return(raw, true)
	mCodec.On("Unmarshal", raw).Return(&cache.Envelope[string]{
		Value: expectedVal,
		DieAt: time.Now().Add(time.Hour).UnixNano(),
	}, nil)

	val, ok, err := mgr.Get(key)

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, expectedVal, val)
}

// Тест удаления при чтении просроченного ключа
func TestManager_Get_Expired(t *testing.T) {
	mStorage := mocks.NewMockStorage[string](t)
	mCodec := mocks.NewMockCodec[string](t)
	mgr := cache.New[string, string](mStorage, mCodec, 0)

	key := "expired-key"
	raw := []byte("old-data")

	mStorage.On("Get", key).Return(raw, true)
	mCodec.On("Unmarshal", raw).Return(&cache.Envelope[string]{
		Value: "some-old-value",
		DieAt: time.Now().Add(-time.Minute).UnixNano(), // Уже просрочено
	}, nil)
	mStorage.On("Delete", key).Return().Once() // Должен сработать триггер удаления

	val, ok, err := mgr.Get(key)

	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, val)
}

// Тест лимита JanitorMaxSize (проверяем, что итератор останавливается)
func TestManager_CacheJanitor_MaxSize(t *testing.T) {
	mStorage := mocks.NewMockStorage[string](t)
	mCodec := mocks.NewMockCodec[string](t)
	// Ставим жесткий лимит: 1 ключ за один проход джанитора
	mgr := cache.New[string, string](mStorage, mCodec, 1)

	now := time.Now()

	// Имитируем, что в сторадже лежат 2 просроченных ключа
	mStorage.On("Range", mock.Anything).Run(func(args mock.Arguments) {
		fn := args.Get(0).(func(string, []byte) bool)
		// Первый пошел
		if !fn("key1", []byte("d1")) {
			return
		}
		// До второго дойти не должен из-за janitorCount >= janitorMaxSize
		fn("key2", []byte("d2"))
	}).Return()

	mCodec.On("Unmarshal", mock.Anything).Return(&cache.Envelope[string]{
		DieAt: now.Add(-time.Minute).UnixNano(),
	}, nil)

	mStorage.On("Delete", "key1").Once()
	// Delete для key2 не должен быть вызван

	err := mgr.CacheJanitor(context.Background(), now)
	assert.NoError(t, err)
}

// Тест прерывания Janitor по контексту (graceful shutdown)
func TestManager_CacheJanitor_ContextCancel(t *testing.T) {
	mStorage := mocks.NewMockStorage[string](t)
	mCodec := mocks.NewMockCodec[string](t)
	mgr := cache.New[string, string](mStorage, mCodec, 100)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем до запуска

	mStorage.On("Range", mock.Anything).Once()

	err := mgr.CacheJanitor(ctx, time.Now())
	assert.ErrorIs(t, err, context.Canceled)
}

// Тест проброса ошибки маршалинга при Set
func TestManager_Set_MarshalError(t *testing.T) {
	mStorage := mocks.NewMockStorage[string](t)
	mCodec := mocks.NewMockCodec[string](t)
	mgr := cache.New[string, string](mStorage, mCodec, 0)

	mCodec.On("Marshal", "val", mock.Anything).Return(nil, errors.New("limit exceeded"))

	err := mgr.Set("key", "val", time.Minute)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marshal failed") // Проверка твоего errs.NewCommonError
}

func TestManager_Integration_JSON(t *testing.T) {
	// Используем реальный кодек, но мокаем сторадж,
	// чтобы не зависеть от реализации памяти/redis
	mStorage := mocks.NewMockStorage[string](t)
	codec := cache.NewJSONCodec[TestData](func() TestData {
		return TestData{}
	})
	mgr := cache.New[string, TestData](mStorage, codec, 10)

	key := "user-123"
	val := TestData{ID: 1, Active: true}
	ttl := time.Hour

	// 1. Тестируем Set
	var capturedBytes []byte
	mStorage.On("Set", key, mock.Anything).Run(func(args mock.Arguments) {
		capturedBytes = args.Get(1).([]byte) // ловим, что кодек выдал в сторадж
	}).Return().Once()

	err := mgr.Set(key, val, ttl)
	assert.NoError(t, err)

	// 2. Тестируем Get (пробрасываем пойманные байты обратно)
	mStorage.On("Get", key).Return(capturedBytes, true).Once()

	res, ok, err := mgr.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, val, res)
}

func TestManager_Integration_Gob(t *testing.T) {
	// Используем реальный кодек, но мокаем сторадж,
	// чтобы не зависеть от реализации памяти/redis
	mStorage := mocks.NewMockStorage[string](t)
	codec := cache.NewGobCodec[TestData](func() TestData {
		return TestData{}
	})
	mgr := cache.New[string, TestData](mStorage, codec, 10)

	key := "user-123"
	val := TestData{ID: 1, Active: true}
	ttl := time.Hour

	// 1. Тестируем Set
	var capturedBytes []byte
	mStorage.On("Set", key, mock.Anything).Run(func(args mock.Arguments) {
		capturedBytes = args.Get(1).([]byte) // ловим, что кодек выдал в сторадж
	}).Return().Once()

	err := mgr.Set(key, val, ttl)
	assert.NoError(t, err)

	// 2. Тестируем Get (пробрасываем пойманные байты обратно)
	mStorage.On("Get", key).Return(capturedBytes, true).Once()

	res, ok, err := mgr.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, val, res)
}

func TestManager_Integration_TTL_Flow(t *testing.T) {
	factory := func() string { return "" }

	// Проверяем на обоих кодеках сразу
	codecs := []struct {
		name  string
		codec cache.Codec[string]
	}{
		{"JSON", cache.NewJSONCodec[string](factory)},
		{"GOB", cache.NewGobCodec[string](factory)},
	}

	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			mStorage := mocks.NewMockStorage[string](t)
			mgr := cache.New[string, string](mStorage, tc.codec, 0)

			key := "ttl-key"
			val := "data"

			// 1. Имитируем Set с TTL в 1 секунду
			var captured []byte
			mStorage.On("Set", key, mock.Anything).Run(func(args mock.Arguments) {
				captured = args.Get(1).([]byte)
			}).Return().Once()

			err := mgr.Set(key, val, time.Second)
			assert.NoError(t, err)

			// 2. Сценарий: Данные еще живы
			mStorage.On("Get", key).Return(captured, true).Once()
			_, ok, err := mgr.Get(key)
			assert.True(t, ok, "Should be alive")
			assert.NoError(t, err)

			// 3. Сценарий: Прошло время, данные просрочены
			// Эмулируем задержку, "подменив" время или просто подождав
			time.Sleep(1100 * time.Millisecond)

			mStorage.On("Get", key).Return(captured, true).Once()
			mStorage.On("Delete", key).Return().Once() // Менеджер должен триггернуть удаление

			_, ok, err = mgr.Get(key)
			assert.False(t, ok, "Should be expired")
			assert.NoError(t, err)
		})
	}
}

func TestCache_FullCycle_Integration(t *testing.T) {
	// 1. Настройка: LRU кэш на 2 элемента с JSON кодеком
	maxSize := 2
	policy := cache.NewLRUEvict[string]()
	storage := cache.NewRawStorage[string](maxSize, policy)

	factory := func() int { return 0 }
	codec := cache.NewJSONCodec[int](factory)

	// janitorMaxSize = 10
	mgr := cache.New[string, int](storage, codec, 10)

	// 2. Проверяем заполнение и TTL
	_ = mgr.Set("a", 100, time.Hour)
	_ = mgr.Set("b", 200, time.Hour)

	val, ok, _ := mgr.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 100, val)

	// 3. Проверяем выселение (Eviction)
	// Сейчас в кэше [a, b] (т.к. к 'a' обращались последним).
	// Добавляем 'c', должно выселиться 'b'.
	_ = mgr.Set("c", 300, time.Hour)

	_, okB, _ := mgr.Get("b")
	assert.False(t, okB, "Элемент 'b' должен быть вытеснен по LRU")

	assert.Equal(t, 2, mgr.Size())

	// 4. Проверяем работу Janitor (TTL)
	// Добавляем элемент с коротким TTL
	_ = mgr.Set("short", 999, time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	// Запускаем очистку вручную
	err := mgr.CacheJanitor(context.Background(), time.Now())
	assert.NoError(t, err)

	_, okShort, _ := mgr.Get("short")
	assert.False(t, okShort, "Элемент с коротким TTL должен быть удален джанитором")
}
