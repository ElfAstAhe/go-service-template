package transport

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestGenerator_Prefix(t *testing.T) {
	t.Run("Проверка формата и структуры префикса инстанса", func(t *testing.T) {
		pfx := GetPrefix()

		assert.NotEmpty(t, pfx)
		assert.Contains(t, pfx, "/")

		parts := strings.Split(pfx, "/")
		assert.Len(t, parts, 2)

		hostname := parts[0]
		salt := parts[1]

		assert.NotEmpty(t, hostname)
		assert.Len(t, salt, 10)
		assert.NotContains(t, salt, "+")
		assert.NotContains(t, salt, "/")
	})
}

func TestRequestGenerator_NextReqID_WithReset(t *testing.T) {
	t.Run("Проверка последовательного инкремента со сбросом счетчика", func(t *testing.T) {
		// Очищаем глобальный счетчик перед тестом через твой хелпер
		testGAPClearReqID()

		id1 := NextReqID()
		id2 := NextReqID()

		assert.Equal(t, uint64(1), id1)
		assert.Equal(t, uint64(2), id2)

		// Сбрасываем снова и проверяем, что отсчет пошел сначала
		testGAPClearReqID()
		id3 := NextReqID()
		assert.Equal(t, uint64(1), id3)
	})
}

func TestRequestGenerator_NextReqID_Concurrency(t *testing.T) {
	t.Run("Конкурентная потокобезопасность под нагрузкой", func(t *testing.T) {
		testGAPClearReqID()

		var wg sync.WaitGroup
		goroutines := 50
		iterationsPerGoroutine := 500
		generatedIDs := sync.Map{}

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterationsPerGoroutine; j++ {
					id := NextReqID()
					_, loaded := generatedIDs.LoadOrStore(id, true)
					if loaded {
						t.Errorf("Обнаружен дубликат ID в конкурентной среде: %d", id)
					}
				}
			}()
		}

		wg.Wait()
	})
}
