package container

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBaseContainer_Lifecycle(t *testing.T) {
	// Создаем мок логгера через mockery
	mockLog := mocks.NewMockLogger(t)

	// Настраиваем логгер, так как конструктор BaseContainer зовет GetLogger
	mockLog.On("GetLogger", "BaseContainer").Return(mockLog)

	c := NewBaseContainer("test-container", mockLog)

	t.Run("Add_And_Get_Success", func(t *testing.T) {
		instance := "hello-world"
		err := c.Add("key1", instance)

		assert.NoError(t, err)
		res, err := c.GetInstance("key1")
		assert.NoError(t, err)
		assert.Equal(t, instance, res)
	})

	t.Run("Add_Duplicate_Error", func(t *testing.T) {
		err := c.Add("key1", "duplicate")
		assert.Error(t, err)
		// Проверяем тип ошибки, если у тебя используется errs.NewContainerError
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Get_NonExistent_Error", func(t *testing.T) {
		res, err := c.GetInstance("unknown")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Remove_Instance", func(t *testing.T) {
		_ = c.Add("to-delete", 123)
		err := c.Remove("to-delete")
		assert.NoError(t, err)

		_, err = c.GetInstance("to-delete")
		assert.Error(t, err)
	})

	t.Run("Validate_Empty_Name", func(t *testing.T) {
		err := c.Add("", "something")
		assert.Error(t, err)
		// Здесь должна сработать твоя commonValidate
		assert.Contains(t, err.Error(), "name is empty")
	})
}

func TestBaseContainer_Concurrency(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", "BaseContainer").Return(mockLog)

	c := NewBaseContainer("concurrency-test", mockLog)

	const workers = 30
	const iterations = 50
	wg := sync.WaitGroup{}
	wg.Add(workers * 2)

	// Параллельная запись
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("w-%d-k-%d", workerID, j)
				_ = c.Add(key, j)
			}
		}(i)
	}

	// Параллельное чтение
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("w-%d-k-%d", workerID, j)
				_, _ = c.GetInstance(key)
			}
		}(i)
	}

	wg.Wait()
	assert.Greater(t, len(c.AllInstances()), 0)
}

func TestBaseContainer_AllInstances_Isolation(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", "BaseContainer").Return(mockLog)

	c := NewBaseContainer("isolation-test", mockLog)
	_ = c.Add("a", 1)

	all := c.AllInstances()
	// Проверяем, что изменение копии не портит оригинал
	all["b"] = 2

	_, err := c.GetInstance("b")
	assert.Error(t, err, "Should not find instance 'b' in container after modifying map copy")
}
