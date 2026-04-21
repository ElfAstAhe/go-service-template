package test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

func TestBaseLazyContainer_LazyInitialization(t *testing.T) {
	ctn := container.NewBaseLazyContainer("test-ctn", nil)

	counter := 0
	instanceName := "lazy-service"

	// Регистрируем провайдер
	err := ctn.RegisterProvider(instanceName, func(name string) (any, error) {
		counter++ // Увеличиваем счетчик при каждом вызове
		return "service-instance", nil
	})

	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// 1. Проверяем, что объект еще не создан
	if counter != 0 {
		t.Errorf("provider should not be called yet, counter: %d", counter)
	}

	// 2. Вызываем GetInstance первый раз
	inst, err := ctn.GetInstance(instanceName)
	if err != nil || inst != "service-instance" {
		t.Errorf("first GetInstance failed: %v", err)
	}
	if counter != 1 {
		t.Errorf("provider should be called once, counter: %d", counter)
	}

	// 3. Вызываем GetInstance второй раз (должен взять из кэша)
	inst2, _ := ctn.GetInstance(instanceName)
	if inst2 != "service-instance" || counter != 1 {
		t.Errorf("provider should NOT be called again, counter: %d", counter)
	}
}

func TestBaseLazyContainer_Concurrency(t *testing.T) {
	ctn := container.NewBaseLazyContainer("race-ctn", nil)

	var createCount int32
	instanceName := "concurrent-service"

	_ = ctn.RegisterProvider(instanceName, func(name string) (any, error) {
		atomic.AddInt32(&createCount, 1)
		time.Sleep(10 * time.Millisecond) // Симулируем тяжелую сборку
		return "ok", nil
	})

	var wg sync.WaitGroup
	const workers = 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = ctn.GetInstance(instanceName)
		}()
	}
	wg.Wait()

	if createCount != 1 {
		t.Errorf("Provider must be called exactly once, but got %d", createCount)
	}
}

func TestBaseLazyContainer_Order(t *testing.T) {
	ctn := container.NewBaseLazyContainer("order-ctn", nil)

	names := []string{"first", "second", "third"}
	for _, name := range names {
		_ = ctn.RegisterProvider(name, func(n string) (any, error) { return "ok", nil })
	}

	// Здесь нужно проверить внутреннее поле order,
	// если оно не экспортировано — можно через AllInstances/AllProviders
	// или просто убедиться, что Unregister правильно чистит слайс.
}

func TestBaseLazyContainer_Unregister(t *testing.T) {
	ctn := container.NewBaseLazyContainer("clean-ctn", nil)
	name := "to-delete"

	_ = ctn.RegisterProvider(name, func(n string) (any, error) { return "val", nil })
	_, _ = ctn.GetInstance(name) // "Будим" объект

	if !ctn.IsRegistered(name) {
		t.Fatal("should be registered")
	}

	_ = ctn.Unregister(name)

	if ctn.IsRegistered(name) {
		t.Error("should be unregistered")
	}

	_, err := ctn.GetInstance(name)
	if err == nil {
		t.Error("GetInstance should return error after Unregister")
	}
}

func TestBaseLazyContainer_ProviderError(t *testing.T) {
	ctn := container.NewBaseLazyContainer("err-ctn", nil)
	name := "flaky-service"

	expectedErrText := "first time fail"
	calls := 0

	_ = ctn.RegisterProvider(name, func(n string) (any, error) {
		calls++
		if calls == 1 {
			return nil, errors.New(expectedErrText)
		}
		return "success", nil
	})

	// 1. Первый вызов — проверяем, что ошибка содержит наш текст
	_, err := ctn.GetInstance(name)
	if err == nil || !strings.Contains(err.Error(), expectedErrText) {
		t.Errorf("expected error containing '%s', got: '%v'", expectedErrText, err)
	}

	// 2. Второй вызов — проверяем, что провайдер вызвался повторно и теперь успешно
	inst, err := ctn.GetInstance(name)
	if err != nil {
		t.Fatalf("expected success on second call, got error: %v", err)
	}

	if inst != "success" {
		t.Errorf("expected 'success', got: %v", inst)
	}

	if calls != 2 {
		t.Errorf("expected provider to be called twice (no caching on error), got %d", calls)
	}
}

func TestBaseLazyContainer_OverrideLogic(t *testing.T) {
	ctn := container.NewBaseLazyContainer("override-ctn", nil)
	name := "service"

	// 1. Сначала регистрируем готовый инстанс
	_ = ctn.RegisterInstance(name, "manual-version")

	// 2. Пытаемся зарегистрировать провайдер с тем же именем
	err := ctn.RegisterProvider(name, func(n string) (any, error) {
		return "lazy-version", nil
	})

	// В текущей логике мы разрешили регистрацию провайдера,
	// но GetInstance должен вернуть то, что уже лежит в мапе instances (priority)
	if err != nil {
		t.Errorf("RegisterProvider should not fail if instance exists: %v", err)
	}

	inst, _ := ctn.GetInstance(name)
	if inst != "manual-version" {
		t.Errorf("expected manual-version to have priority, got: %v", inst)
	}
}

func TestBaseLazyContainer_SelfDependency(t *testing.T) {
	ctn := container.NewBaseLazyContainer("deadlock-ctn", nil)

	// Регистрируем А, который зависит от Б
	_ = ctn.RegisterProvider("A", func(n string) (any, error) {
		depB, err := ctn.GetInstance("B")
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("A depends on %s", depB), nil
	})

	// Регистрируем Б
	_ = ctn.RegisterProvider("B", func(n string) (any, error) {
		return "B-instance", nil
	})

	// Вызов GetInstance("A") не должен повесить систему (deadlock)
	res, err := ctn.GetInstance("A")
	if err != nil {
		t.Fatalf("failed to get A: %v", err)
	}

	expected := "A depends on B-instance"
	if res != expected {
		t.Errorf("expected %s, got %s", expected, res)
	}
}

func TestBaseLazyContainer_AllProvidersConsistency(t *testing.T) {
	ctn := container.NewBaseLazyContainer("list-ctn", nil)

	_ = ctn.RegisterProvider("p1", func(n string) (any, error) { return 1, nil })
	_ = ctn.RegisterProvider("p2", func(n string) (any, error) { return 2, nil })

	providers := ctn.AllProviders()
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}

	// "Будим" один инстанс
	_, _ = ctn.GetInstance("p1")

	// Список провайдеров не должен измениться
	providersAfter := ctn.AllProviders()
	if len(providersAfter) != 2 {
		t.Errorf("expected 2 providers after instance creation, got %d", len(providersAfter))
	}
}

func TestBaseLazyContainer_InterfaceCompliance(t *testing.T) {
	var _ container.Container = (*container.BaseLazyContainer)(nil)
	var _ container.LazyContainer = (*container.BaseLazyContainer)(nil)
}
