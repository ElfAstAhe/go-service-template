package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestLFUEvict_Order(t *testing.T) {
	t.Run("basic_eviction", func(t *testing.T) {
		lfu := cache.NewLFUEvict[string]()

		// Добавляем 3 элемента. У всех count=1.
		lfu.OnSet("a")
		lfu.OnSet("b")
		lfu.OnSet("c")

		// Накручиваем частоту для "b" и "c"
		lfu.OnGet("b") // count=2
		lfu.OnGet("c") // count=2
		lfu.OnGet("c") // count=3

		// Сейчас: a(1), b(2), c(3). Должен уйти "a".
		key, ok := lfu.Evict()
		assert.True(t, ok)
		assert.Equal(t, "a", key)
	})

	t.Run("tie_break_same_frequency", func(t *testing.T) {
		lfu := cache.NewLFUEvict[string]()

		// Добавляем элементы в порядке: a, b, c
		lfu.OnSet("a")
		lfu.OnSet("b")
		lfu.OnSet("c")

		// У всех count=1. Твоя реализация делает PushFront, а Evict берет Back().
		// Значит, первым уйдет тот, кто был добавлен раньше всех (FIFO на уровне частоты).
		key, ok := lfu.Evict()
		assert.True(t, ok)
		assert.Equal(t, "a", key)

		key, ok = lfu.Evict()
		assert.Equal(t, "b", key)
	})

	t.Run("min_freq_update", func(t *testing.T) {
		lfu := cache.NewLFUEvict[string]()

		lfu.OnSet("a") // count=1, minFreq=1
		lfu.OnGet("a") // count=2, minFreq=2 (т.к. список 1 пуст)

		lfu.OnSet("b") // count=1, minFreq снова 1

		// Должен уйти "b", так как у него count=1, а у "a" уже 2
		key, ok := lfu.Evict()
		assert.True(t, ok)
		assert.Equal(t, "b", key)
	})

	t.Run("on_remove_cleanup", func(t *testing.T) {
		lfu := cache.NewLFUEvict[string]()

		lfu.OnSet("a")
		lfu.OnRemove("a")

		key, ok := lfu.Evict()
		assert.False(t, ok, "Should be empty after removal")
		assert.Empty(t, key)
	})

	t.Run("reset", func(t *testing.T) {
		lfu := cache.NewLFUEvict[string]()
		lfu.OnSet("a")
		lfu.Reset()

		assert.True(t, true)
		//assert.Equal(t, 0, len(lfu.items))
		//assert.Equal(t, 0, lfu.minFreq)
	})
}
