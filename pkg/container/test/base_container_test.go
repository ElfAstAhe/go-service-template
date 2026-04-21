package test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/stretchr/testify/assert"
)

func TestBaseContainer_Lifecycle(t *testing.T) {
	c := container.NewBaseContainer("test-container", nil)

	t.Run("Add_And_Get_Success", func(t *testing.T) {
		instance := "hello-world"
		err := c.RegisterInstance("key1", instance)

		assert.NoError(t, err)
		res, err := c.GetInstance("key1")
		assert.NoError(t, err)
		assert.Equal(t, instance, res)
	})

	t.Run("Add_Duplicate_Error", func(t *testing.T) {
		err := c.RegisterInstance("key1", "duplicate")
		assert.Error(t, err)
		// Проверяем тип ошибки, если у тебя используется errs.NewContainerError
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Get_NonExistent_Error", func(t *testing.T) {
		res, err := c.GetInstance("unknown")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "not registered")
	})

	t.Run("Remove_Instance", func(t *testing.T) {
		_ = c.RegisterInstance("to-delete", 123)
		err := c.UnregisterInstance("to-delete")
		assert.NoError(t, err)

		_, err = c.GetInstance("to-delete")
		assert.Error(t, err)
	})

	t.Run("Validate_Empty_Name", func(t *testing.T) {
		err := c.RegisterInstance("", "something")
		assert.Error(t, err)
		// Здесь должна сработать твоя commonValidate
		assert.Contains(t, err.Error(), "name is empty")
	})
}

func TestBaseContainer_Concurrency(t *testing.T) {
	c := container.NewBaseContainer("concurrency-test", nil)

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
				_ = c.RegisterInstance(key, j)
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
	assert.Greater(t, len(c.AllNames()), 0)
}
