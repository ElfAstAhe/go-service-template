package container

import (
	"fmt"
	"slices"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type BaseLazyContainer struct {
	*BaseContainer
	mu    sync.RWMutex
	names map[string]struct{}
	// Карта каналов-обещаний: если ключ есть, значит объект в процессе создания
	inProgress map[string]chan struct{}
	order      []string
	providers  map[string]Provider
}

func NewBaseLazyContainer(name string) *BaseLazyContainer {
	return &BaseLazyContainer{
		BaseContainer: NewBaseContainer(name),
		names:         make(map[string]struct{}),
		order:         make([]string, 0),
		providers:     make(map[string]Provider),
		inProgress:    make(map[string]chan struct{}),
	}
}

var _ Container = (*BaseLazyContainer)(nil)
var _ LazyContainer = (*BaseLazyContainer)(nil)

func (blc *BaseLazyContainer) GetInstance(name string) (any, error) {
	// 1. Быстрая проверка: вдруг уже создано?
	if res, err := blc.BaseContainer.GetInstance(name); err == nil {
		return res, nil
	}

	blc.mu.Lock()
	// 2. Double-check под локом
	if res, err := blc.BaseContainer.GetInstance(name); err == nil {
		blc.mu.Unlock()
		return res, nil
	}

	// 3. Проверяем "обещание" (Promise)
	if waiter, found := blc.inProgress[name]; found {
		blc.mu.Unlock()
		<-waiter // Ждем, пока первый поток закончит
		return blc.BaseContainer.GetInstance(name)
	}

	// 4. Мы — "Первопроходцы".
	ch := make(chan struct{})
	blc.inProgress[name] = ch

	provider, ok := blc.providers[name]
	blc.mu.Unlock() // ОТПУСКАЕМ ГЛОБАЛЬНЫЙ ЛОК

	// Гарантируем очистку канала при любом исходе (паника, ошибка, успех)
	defer blc.notifyAndCleanup(name, ch)

	if !ok {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("provider [%s] not registered", name), nil)
	}

	// 5. Спокойно создаем объект ВНЕ лока.
	res, err := provider(name)
	if err != nil {
		return nil, err
	}

	// 6. Регистрируем готовый результат (внутри BaseContainer свой Lock)
	if regErr := blc.BaseContainer.RegisterInstance(name, res); regErr != nil {
		return nil, regErr
	}

	return res, nil
}

// notifyAndCleanup — вспомогательный приватный метод
func (blc *BaseLazyContainer) notifyAndCleanup(name string, ch chan struct{}) {
	blc.mu.Lock()
	defer blc.mu.Unlock()

	delete(blc.inProgress, name)
	close(ch) // Сигнал всем "ждунам"
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
