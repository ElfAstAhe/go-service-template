package test

import (
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestL2Manager_NegativeCaching(t *testing.T) {
	// 1. Инициализируем L2Manager через фабрику (или напрямую)
	// Используем JSON кодек для простоты
	factory := func() string { return "" }
	codec := cache.NewJSONCodec[string](factory)
	storage := cache.NewRawStorage[string](10, cache.NewLRUEvict[string]())

	l2 := cache.NewL2[string, string](storage, codec, 10)

	key := "non-existent-key"

	// 2. Первый вызов Get — ключа точно нет
	val1, ok1, err1 := l2.Get(key)

	assert.NoError(t, err1)
	assert.False(t, ok1, "First call should return ok=false")
	assert.Empty(t, val1)

	// 3. А теперь магия: L2Manager должен был сам сделать Set
	// Проверяем это через Has в сторадже или просто вторым вызовом Get
	// Но есть нюанс: Get в L2Manager всё еще возвращает ok=false,
	// если он вычитал "нулевое значение".

	// Проверим, что в сторадже появились байты по этому ключу
	assert.True(t, storage.Has(key), "L2Manager should put negative value into storage")

	// 4. Проверяем, что данные реально лежат в кэше (через базовый Manager)
	// Если вызвать напрямую Manager.Get, он должен вернуть ok=true для "пустоты"
	val2, ok2, err2 := l2.Manager.Get(key)
	assert.NoError(t, err2)
	assert.True(t, ok2, "Base Manager should see the negative cached value as ok=true")
	assert.Empty(t, val2)
}

func TestL2Manager_NormalFlow(t *testing.T) {
	// Проверяем, что обычное сохранение работает как надо
	factory := func() int { return 0 }
	codec := cache.NewJSONCodec[int](factory)
	storage := cache.NewRawStorage[string](10, cache.NewLRUEvict[string]())
	l2 := cache.NewL2[string, int](storage, codec, 10)

	key := "lucky-number"
	val := 777

	// Сохраняем нормально
	err := l2.Set(key, val, time.Hour)
	assert.NoError(t, err)

	// Читаем — должно быть ok=true
	res, ok, err := l2.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, val, res)
}
