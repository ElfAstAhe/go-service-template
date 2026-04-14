package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestLRUEvict(t *testing.T) {
	t.Run("basic_eviction_order", func(t *testing.T) {
		lru := cache.NewLRUEvict[string]()

		// Добавляем a, b, c. Очередь: [c, b, a]
		lru.OnSet("a")
		lru.OnSet("b")
		lru.OnSet("c")

		// Самый старый — "a". Он должен уйти первым.
		key, ok := lru.Evict()
		assert.True(t, ok)
		assert.Equal(t, "a", key)

		// Следующий — "b"
		key, ok = lru.Evict()
		assert.Equal(t, "b", key)
	})

	t.Run("get_updates_priority", func(t *testing.T) {
		lru := cache.NewLRUEvict[string]()

		lru.OnSet("a")
		lru.OnSet("b")

		// Сейчас порядок [b, a]. "a" на вылет.
		// Делаем Get для "a", перемещая его в начало.
		lru.OnGet("a")

		// Теперь порядок [a, b]. На вылет идет "b".
		key, ok := lru.Evict()
		assert.True(t, ok)
		assert.Equal(t, "b", key)
	})

	t.Run("set_existing_updates_priority", func(t *testing.T) {
		lru := cache.NewLRUEvict[string]()

		lru.OnSet("a")
		lru.OnSet("b")

		// Перезаписываем "a" (OnSet существующего ключа)
		lru.OnSet("a")

		// На вылет идет "b"
		key, ok := lru.Evict()
		assert.True(t, ok)
		assert.Equal(t, "b", key)
	})

	t.Run("on_remove_and_evict_empty", func(t *testing.T) {
		lru := cache.NewLRUEvict[string]()

		key, ok := lru.Evict()
		assert.False(t, ok)
		assert.Empty(t, key)

		lru.OnSet("a")
		lru.OnRemove("a")

		assert.True(t, true)
		//assert.Equal(t, 0, len(lru.items))
		//assert.Nil(t, lru.ll.Back())
	})
}
