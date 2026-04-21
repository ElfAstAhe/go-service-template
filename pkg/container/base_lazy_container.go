package container

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type BaseLazyContainer struct {
	*BaseContainer
	mu        sync.RWMutex
	names     map[string]struct{}
	order     []string
	providers map[string]Provider
}

func NewBaseLazyContainer(name string) *BaseLazyContainer {
	return &BaseLazyContainer{
		BaseContainer: NewBaseContainer(name),
		names:         make(map[string]struct{}),
		order:         make([]string, 0),
		providers:     make(map[string]Provider),
	}
}

var _ Container = (*BaseLazyContainer)(nil)
var _ LazyContainer = (*BaseLazyContainer)(nil)

func (blc *BaseLazyContainer) GetInstance(name string) (any, error) {
	if err := blc.Validate("BaseLazyContainer.GetInstance", name); err != nil {
		return nil, err
	}
	if !blc.IsRegistered(name) {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("instance or provider [%s] not registered", name), nil)
	}

	// 1. Быстрая проверка (RLock внутри базы)
	res, instanceErr := blc.BaseContainer.GetInstance(name)
	if instanceErr == nil {
		return res, nil
	}

	// 2. Получаем рецепт (RLock внутри)
	provider, providerErr := blc.getProvider(name)
	if providerErr != nil {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("instance and provider [%s] not registered", name), errors.Join(instanceErr, providerErr))
	}

	// 3. ЗАХВАТЫВАЕМ LOCK НА ВЕСЬ ПРОЦЕСС СОЗДАНИЯ
	blc.mu.Lock()
	defer blc.mu.Unlock()

	// 4. Double-Check: вдруг пока мы ждали Lock, кто-то уже создал объект
	if res, err := blc.BaseContainer.GetInstance(name); err == nil {
		return res, nil
	}

	// 5. Теперь мы точно единственные, кто создает объект
	res, instanceErr = provider(name)
	if instanceErr != nil {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("create instance [%s] failed", name), instanceErr)
	}

	// 6. Регистрация
	instanceErr = blc.BaseContainer.RegisterInstance(name, res)
	if instanceErr != nil {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("register instance [%s] failed", name), instanceErr)
	}

	blc.names[name] = struct{}{}

	return res, nil
}

func (blc *BaseLazyContainer) RegisterProvider(name string, provider Provider) error {
	if err := blc.Validate("BaseLazyContainer.RegisterProvider", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// check existence
	if _, ok := blc.providers[name]; ok {
		return errs.NewContainerError(blc.GetName(), fmt.Sprintf("provider %s already registered", name), nil)
	}

	blc.providers[name] = provider
	if !blc.BaseContainer.IsRegistered(name) {
		blc.order = append(blc.order, name)
	}
	blc.names[name] = struct{}{}

	return nil
}

func (blc *BaseLazyContainer) UnregisterProvider(name string) error {
	if err := blc.Validate("BaseLazyContainer.UnregisterProvider", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	delete(blc.providers, name)
	if !blc.BaseContainer.IsRegistered(name) {
		delete(blc.names, name)
		blc.order = slices.DeleteFunc(blc.order, func(item string) bool {
			return item == name
		})
	}

	return nil
}

func (blc *BaseLazyContainer) AllProviders() map[string]Provider {
	blc.mu.RLock()
	defer blc.mu.RUnlock()

	res := make(map[string]Provider, len(blc.providers))
	for k, v := range blc.providers {
		res[k] = v
	}

	return res
}

func (blc *BaseLazyContainer) IsRegistered(name string) bool {
	blc.mu.RLock()
	defer blc.mu.RUnlock()

	_, ok := blc.names[name]

	return ok
}

// Unregister remove provider and instance from lists, errors ignored
func (blc *BaseLazyContainer) Unregister(name string) error {
	if err := blc.Validate("BaseLazyContainer.Unregister", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// remove from provider list
	delete(blc.providers, name)
	// remove from instance list
	_ = blc.BaseContainer.UnregisterInstance(name)

	delete(blc.names, name)
	blc.order = slices.DeleteFunc(blc.order, func(item string) bool {
		return item == name
	})

	return nil
}

func (blc *BaseLazyContainer) getProvider(name string) (Provider, error) {
	if !blc.isProviderRegistered(name) {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("provider [%s] not registered", name), nil)
	}

	blc.mu.RLock()
	defer blc.mu.RUnlock()

	provider, ok := blc.providers[name]
	if !ok {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("provider [%s] not registered", name), nil)
	}

	return provider, nil
}

func (blc *BaseLazyContainer) isProviderRegistered(name string) bool {
	blc.mu.RLock()
	defer blc.mu.RUnlock()

	_, ok := blc.providers[name]

	return ok
}

func (blc *BaseLazyContainer) RegisterInstance(name string, instance any) error {
	if err := blc.Validate("BaseLazyContainer.RegisterInstance", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// 1. Пытаемся сохранить в базовый склад
	if err := blc.BaseContainer.RegisterInstance(name, instance); err != nil {
		return err
	}

	// 2. Если база приняла (не дубликат), фиксируем в нашем реестре
	if _, ok := blc.names[name]; !ok {
		blc.names[name] = struct{}{}
		blc.order = append(blc.order, name)
	}

	return nil
}

func (blc *BaseLazyContainer) UnregisterInstance(name string) error {
	if err := blc.Validate("BaseLazyContainer.UnregisterInstance", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// 1. Удаляем из базы
	if err := blc.BaseContainer.UnregisterInstance(name); err != nil {
		return err
	}

	// 2. Если объекта нет и в провайдерах — вычищаем из порядка
	if _, isProvider := blc.providers[name]; !isProvider {
		delete(blc.names, name)
		blc.order = slices.DeleteFunc(blc.order, func(item string) bool {
			return item == name
		})
	}

	return nil
}
