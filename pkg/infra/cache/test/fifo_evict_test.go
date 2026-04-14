package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestFIFOEvict(t *testing.T) {
	t.Run("strict_order", func(t *testing.T) {
		fifo := cache.NewFIFOEvict[string]()

		// Добавляем в порядке a, b, c
		fifo.OnSet("a")
		fifo.OnSet("b")
		fifo.OnSet("c")

		// Первым зашел — первым вышел (a)
		key, ok := fifo.Evict()
		assert.True(t, ok)
		assert.Equal(t, "a", key)

		// Вторым зашел — вторым вышел (b)
		key, ok = fifo.Evict()
		assert.True(t, ok)
		assert.Equal(t, "b", key)
	})

	t.Run("get_does_not_change_order", func(t *testing.T) {
		fifo := cache.NewFIFOEvict[string]()

		fifo.OnSet("a")
		fifo.OnSet("b")

		// Читаем "a". В LRU это бы его спасло, в FIFO — нет.
		fifo.OnGet("a")

		// "a" всё равно первый в очереди на вылет
		key, ok := fifo.Evict()
		assert.True(t, ok)
		assert.Equal(t, "a", key)
	})

	t.Run("set_existing_does_not_change_order", func(t *testing.T) {
		fifo := cache.NewFIFOEvict[string]()

		fifo.OnSet("a")
		fifo.OnSet("b")

		// Повторный OnSet для "a"
		fifo.OnSet("a")

		// Порядок не изменился, "a" всё еще старее, чем "b"
		key, ok := fifo.Evict()
		assert.True(t, ok)
		assert.Equal(t, "a", key)
	})

	t.Run("reset_clears_everything", func(t *testing.T) {
		fifo := cache.NewFIFOEvict[string]()
		fifo.OnSet("a")
		fifo.Reset()

		key, ok := fifo.Evict()
		assert.False(t, ok)
		assert.Empty(t, key)
	})
}
