package container

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseLazyContainer реализует интерфейс LazyContainer, расширяя базовый BaseContainer.
//
// Инкапсулирует механизмы отложенной инициализации (Lazy Loading) и многопоточной синхронизации
// на базе паттерна Single Flight (каналов-обещаний Channel Promises). Это гарантирует строго
// однократный вызов тяжелых конструкторов зависимостей в конкурентной Highload-среде.
type BaseLazyContainer struct {
	*BaseContainer                          // Встраиваем базовый контейнер для управления кэшированными инстансами
	mu             sync.RWMutex             // RWMutex для защиты локального реестра провайдеров и карты обещаний
	names          map[string]struct{}      // Индексированная карта всех известных контейнеру имен (провайдеров и инстансов)
	inProgress     map[string]chan struct{} // Реестр активных фоновых потоков сборки (Channel Promises) для защиты от Thundering Herd
	order          []string                 // Хронологический слайс имен в порядке регистрации для соблюдения этапов деструкции
	providers      map[string]Provider      // Реестр легковесных замыканий-фабрик (конструкторов зависимостей)
	logger         logger.Logger            // Изолированный структурированный логгер ленивого контейнера
}

// Гарантируем полное соответствие интерфейсам контейнеров на этапе компиляции
var _ Container = (*BaseLazyContainer)(nil)
var _ LazyContainer = (*BaseLazyContainer)(nil)

// NewBaseLazyContainer — фабричный конструктор базового ленивого контейнера зависимостей.
func NewBaseLazyContainer(
	opts ...LazyOption,
) *BaseLazyContainer {
	lazyOptions := &LazyOptions{}

	// Вычисляем входящие мутаторы Fluent API
	for _, o := range opts {
		o(lazyOptions)
	}

	return &BaseLazyContainer{
		BaseContainer: NewBaseContainer(
			WithName(lazyOptions.Name),
			WithOrchestrator(lazyOptions.Orchestrator),
			WithLogger(lazyOptions.Logger),
		),
		names:      make(map[string]struct{}),
		order:      make([]string, 0),
		providers:  make(map[string]Provider),
		inProgress: make(map[string]chan struct{}),
		logger:     lazyOptions.Logger.GetLogger("BaseLazyContainer"),
	}
}

// GetInstance осуществляет строго однократное, потокобезопасное разрешение зависимости.
// Если инстанс еще не создан, метод блокирует конкурентные потоки-ждуны до завершения сборки «первопроходцем».
func (blc *BaseLazyContainer) GetInstance(name string) (any, error) {
	blc.logger.Debugf("container %s get instance: started", blc.GetName())
	defer blc.logger.Debugf("container %s get instance: finished", blc.GetName())

	// Шаг 1. Быстрый проход (Fast-Path): вдруг инстанс уже был собран и кэширован ранее?
	if res, err := blc.BaseContainer.GetInstance(name); err == nil {
		return res, nil
	}

	blc.mu.Lock()
	// Шаг 2. Двойная проверка (Double-Checked Locking) под эксклюзивным локом
	if res, err := blc.BaseContainer.GetInstance(name); err == nil {
		blc.mu.Unlock()
		return res, nil
	}

	// Шаг 3. Анализ обещаний (Promise Verification): если компонент уже собирается другой горутиной
	if waiter, found := blc.inProgress[name]; found {
		blc.mu.Unlock()                            // Мгновенно отпускаем лок, давая дорогу другим читателям
		<-waiter                                   // Элегантная неблокирующая CPU блокировка потока до момента закрытия канала (close)
		return blc.BaseContainer.GetInstance(name) // Возвращаем результат, собранный соседом
	}

	// Шаг 4. Статус «Первопроходца»: текущий поток берет на себя обязательство по вызову фабрики.
	ch := make(chan struct{})
	blc.inProgress[name] = ch

	provider, ok := blc.providers[name]
	blc.mu.Unlock() // 🧠 ВАЖНО: Отпускаем глобальный лок до старта тяжелых операций ввода-вывода (I/O)

	// Намертво страхуем систему дефером: канал обязан закрыться, даже если фабрика выкинет panic()
	defer blc.notifyAndCleanup(name, ch)

	if !ok {
		return nil, errs.NewContainerError(blc.GetName(), fmt.Sprintf("provider [%s] not registered", name), nil)
	}

	// Шаг 5. Спокойно и изолированно исполняем конструктор ВНЕ мьютекса
	res, err := provider()
	if err != nil {
		return nil, err
	}

	// Шаг 6. Фиксируем готовый результат в кэше базового контейнера (там отработает свой локальный Lock)
	if regErr := blc.RegisterInstance(name, res); regErr != nil {
		return nil, regErr
	}

	return res, nil
}

// notifyAndCleanup атомарно вырезает обещание из карты рантайма и закрывает канал,
// пробуждая и высвобождая все заблокированные горутины-ждуны (Fan-Out Broadcast).
func (blc *BaseLazyContainer) notifyAndCleanup(name string, ch chan struct{}) {
	blc.mu.Lock()
	defer blc.mu.Unlock()

	delete(blc.inProgress, name)
	close(ch) // Сигнальный маркер для всех "ждунов"
}

// RegisterProvider регистрирует пассивную фабрику-замыкание во внутреннем словаре провайдеров.
func (blc *BaseLazyContainer) RegisterProvider(name string, provider Provider) error {
	blc.logger.Debugf("lazy container %s register provider: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s register provider: finished", blc.GetName())

	if err := blc.Validate("BaseLazyContainer.RegisterProvider", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

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

// RegisterRunnableProvider — заглушка для будущей регистрации воркеров и серверов (Runners).
func (blc *BaseLazyContainer) RegisterRunnableProvider(name string, provider Provider) error {
	blc.logger.Debugf("lazy container %s register runnable provider: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s register runnable provider: finished", blc.GetName())

	if err := blc.Validate("BaseLazyContainer.RegisterRunnableProvider", name); err != nil {
		return err
	}

	// ToDo: implement
	return nil
}

// UnregisterProvider принудительно удаляет фабрику из реестра с перестройкой среза хронологии порядка.
func (blc *BaseLazyContainer) UnregisterProvider(name string) error {
	blc.logger.Debugf("lazy container %s unregister provider: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s unregister provider: finished", blc.GetName())

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

// AllProviders экспортирует изолированный слепок текущей карты фабрик-провайдеров.
func (blc *BaseLazyContainer) AllProviders() map[string]Provider {
	blc.logger.Debugf("lazy container %s all providers: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s all providers: finished", blc.GetName())

	blc.mu.RLock()
	defer blc.mu.RUnlock()

	res := make(map[string]Provider, len(blc.providers))
	for k, v := range blc.providers {
		res[k] = v
	}

	return res
}

// AllNames возвращает полный хронологический массив идентификаторов всех известных ленивому контейнеру сущностей.
func (blc *BaseLazyContainer) AllNames() []string {
	blc.logger.Debugf("lazy container %s all names: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s all names: finished", blc.GetName())

	blc.mu.RLock()
	defer blc.mu.RUnlock()

	res := make([]string, len(blc.order))
	copy(res, blc.order)

	return res
}

// IsRegistered проверяет под RLock-блокировкой, существует ли зарегистрированное имя
// (будь то активный инстанс или еще не вызванный провайдер) в реестре ленивого контейнера.
func (blc *BaseLazyContainer) IsRegistered(name string) bool {
	blc.logger.Debugf("lazy container %s is registered: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s is registered: finished", blc.GetName())

	blc.mu.RLock()
	defer blc.mu.RUnlock()

	_, ok := blc.names[name]

	return ok
}

// Unregister каскадно вырезает из памяти как декларативную фабрику-провайдер,
// так и живой физический объект зависимости, полностью игнорируя промежуточные ошибки.
func (blc *BaseLazyContainer) Unregister(name string) error {
	blc.logger.Debugf("lazy container %s unregister: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s unregister: finished", blc.GetName())

	if err := blc.Validate("BaseLazyContainer.Unregister", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// Вырезаем фабрику-провайдер из реестра отложенной инициализации
	delete(blc.providers, name)
	// Принудительно выгружаем созданный объект из базового хранилища (ошибки подавляются)
	_ = blc.BaseContainer.UnregisterInstance(name)

	delete(blc.names, name)
	// Схлопываем хронологический слайс порядка без лишних переаллокаций в куче
	blc.order = slices.DeleteFunc(blc.order, func(item string) bool {
		return item == name
	})

	return nil
}

// getProvider извлекает зарегистрированное замыкание-конструктор по его имени.
//
//lint:ignore U1000 This method is kept for future extensions or interface compatibility
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

// isProviderRegistered осуществляет атомарную проверку наличия фабрики в реестре провайдеров.
//
//lint:ignore U1000 This method is kept for future extensions or interface compatibility
func (blc *BaseLazyContainer) isProviderRegistered(name string) bool {
	blc.mu.RLock()
	defer blc.mu.RUnlock()

	_, ok := blc.providers[name]

	return ok
}

// RegisterInstance расширяет базовый метод фиксации объектов, параллельно обновляя хронологию ленивого графа.
func (blc *BaseLazyContainer) RegisterInstance(name string, instance any) error {
	blc.logger.Debugf("lazy container %s register instance: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s register instance: finished", blc.GetName())

	if err := blc.Validate("BaseLazyContainer.RegisterInstance", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// 1. Пытаемся сохранить объект во внутренний хэш-склад базового контейнера
	if err := blc.BaseContainer.RegisterInstance(name, instance); err != nil {
		return err
	}

	// 2. Если базовая мапа приняла объект (нет коллизий) — фиксируем его в хронологии деструкции
	if _, ok := blc.names[name]; !ok {
		blc.names[name] = struct{}{}
		blc.order = append(blc.order, name)
	}

	return nil
}

// UnregisterInstance выгружает объект из базового хранилища с синхронизацией локальных структур индексов.
func (blc *BaseLazyContainer) UnregisterInstance(name string) error {
	blc.logger.Debugf("lazy container %s unregister instance: started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s unregister instance: finished", blc.GetName())

	if err := blc.Validate("BaseLazyContainer.UnregisterInstance", name); err != nil {
		return err
	}

	blc.mu.Lock()
	defer blc.mu.Unlock()

	// 1. Вырезаем физический инстанс из базовой мапы памяти
	if err := blc.BaseContainer.UnregisterInstance(name); err != nil {
		return err
	}

	// 2. Если объекта больше нет и в качестве фабрики-провайдера — полностью стираем его упоминание
	if _, isProvider := blc.providers[name]; !isProvider {
		delete(blc.names, name)
		blc.order = slices.DeleteFunc(blc.order, func(item string) bool {
			return item == name
		})
	}

	return nil
}

// Close перехватывает процедуру Graceful Shutdown, параллельно и безопасно гася все лениво созданные инстансы.
func (blc *BaseLazyContainer) Close(closeCtx context.Context) error {
	blc.logger.Debugf("lazy container %s close started", blc.GetName())
	defer blc.logger.Debugf("lazy container %s close finished", blc.GetName())

	var wg sync.WaitGroup

	// 1. Блокируем и собираем созданные инстансы СТРОГО в хронологическом порядке их регистрации (blc.order)
	blc.mu.RLock()
	toClosing := make([]*closeInstance, 0, len(blc.order))
	for _, name := range blc.order {
		if inst, err := blc.BaseContainer.GetInstance(name); err == nil {
			toClosing = append(toClosing, newCloseInstance(name, inst))
		}
	}
	blc.mu.RUnlock()

	blc.logger.Debugf("lazy container %s: got %d instances to close", blc.GetName(), len(toClosing))

	// 2. Полностью обнуляем локальные реестры и карты под эксклюзивным Lock
	blc.mu.Lock()
	blc.providers = make(map[string]Provider)
	blc.names = make(map[string]struct{})
	blc.order = make([]string, 0)
	blc.mu.Unlock()

	closeChan := make(chan struct{})
	closeErrs := utils.NewConcurrentList[error]()

	// 3. Запускаем конкурентный веерный цикл параллельной деструкции ресурсов (Fan-Out Cleanup)
	for _, toClose := range toClosing {
		if inst, ok := toClose.Instance.(SimpleCloser); ok {
			wg.Add(1)
			go func(name string, closer SimpleCloser) {
				blc.logger.Debugf("lazy container %s close: simple closer for instance %s start", blc.GetName(), name)
				defer blc.logger.Debugf("lazy container %s close: simple closer for instance %s finish", blc.GetName(), name)
				defer wg.Done()

				if err := closer.Close(); err != nil {
					closeErrs.Append(err)
					blc.logger.Debugf("container %s close: simple closer for instance %s failed: %v", blc.GetName(), name, err)
				} else {
					blc.logger.Debugf("container %s close: simple closer for instance %s done", blc.GetName(), name)
				}
			}(toClose.Name, inst)
		} else if inst, ok := toClose.Instance.(ContextCloser); ok {
			wg.Add(1)
			go func(name string, closer ContextCloser) {
				blc.logger.Debugf("lazy container %s close: context closer for instance %s start", blc.GetName(), name)
				defer blc.logger.Debugf("lazy container %s close: context closer for instance %s finish", blc.GetName(), name)
				defer wg.Done()

				if err := closer.Close(closeCtx); err != nil {
					closeErrs.Append(err)
					blc.logger.Debugf("lazy container %s close: context closer for instance %s failed: %v", blc.GetName(), name, err)
				} else {
					blc.logger.Debugf("lazy container %s close: context closer for instance %s done", blc.GetName(), name)
				}
			}(toClose.Name, inst)
		} else {
			blc.logger.Debugf("lazy container %s close: no closer methods for instance %s", blc.GetName(), toClose.Name)
		}
	}

	// Селекторный барьер контроля жесткого тайм-аута остановки Kubernetes
	go func() {
		defer close(closeChan)
		wg.Wait()
	}()

	select {
	case <-closeChan:
		if closeErrs.Len() > 0 {
			return errs.NewContainerError(blc.GetName(), "lazy container close: close fails", errors.Join(closeErrs.Snapshot()...))
		}
		return nil
	case <-closeCtx.Done():
		return errs.NewContainerError(blc.GetName(), "lazy container close: close timeout limit reached", closeCtx.Err())
	}
}
