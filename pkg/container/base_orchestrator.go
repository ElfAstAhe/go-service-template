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

// BaseOrchestrator реализует интерфейс Orchestrator, являясь центральным ядром IoC/DI подсистемы фреймворка.
//
// Управляет детерминированным графом инфраструктурных модулей (Containers), отслеживает строгий хронологический
// порядок их регистрации для обеспечения предсказуемой инициализации (FIFO) и каскадного тушения ресурсов (LIFO).
type BaseOrchestrator struct {
	mu       sync.RWMutex         // RWMutex защищает внутренний реестр и слайс порядка от Race Condition в рантайме
	items    map[string]Container // Карта быстрого O(1) доступа к зарегистрированным контейнерам по их именам
	regOrder []string             // Хронологический слайс имен контейнеров для строгого соблюдения этапов жизненного цикла
	log      logger.Logger        // Изолированный структурированный логгер оркестратора
}

// NewBaseOrchestrator — фабричный конструктор базового оркестратора зависимостей.
func NewBaseOrchestrator(log logger.Logger) *BaseOrchestrator {
	return &BaseOrchestrator{
		log:      log.GetLogger("BaseOrchestrator"),
		items:    make(map[string]Container),
		regOrder: make([]string, 0),
	}
}

// Гарантируем полное соответствие интерфейсу Orchestrator на этапе компиляции
var _ Orchestrator = (*BaseOrchestrator)(nil)

// Init последовательно инициализирует (FIFO) все зарегистрированные контейнеры под защитой RLock-блокировки.
func (o *BaseOrchestrator) Init(ctx context.Context) error {
	o.log.Debug("Init start")
	defer o.log.Debug("Init finish")

	o.mu.RLock()
	defer o.mu.RUnlock()

	// Инициализируем контейнеры строго в порядке их добавления в систему (First-In, First-Init)
	for _, name := range o.regOrder {
		ctn := o.items[name]
		o.log.Debugf("initializing container [%s]...", name)
		if err := ctn.Init(ctx); err != nil {
			return err // Прерываем старт при сбое инициализации любого базового компонента (fail-fast)
		}
	}

	return nil
}

// Close каскадно закрывает (LIFO) все контейнеры, минимизируя блокировки и собирая ошибки закрытия через errors.Join.
func (o *BaseOrchestrator) Close(ctx context.Context) error {
	o.log.Debug("Close start")
	defer o.log.Debug("Close finish")

	o.mu.RLock()
	defer o.mu.RUnlock()

	var closeErrs []error
	// 💡 Архитектурный паттерн (LIFO Shutdown): идем по слайсу имен с конца (Last-In, First-Close).
	// Защищает дочерние контейнеры от падения, гася сначала их, а затем — тяжелые родительские пулы СУБД/Кафки.
	for i := len(o.regOrder) - 1; i >= 0; i-- {
		name := o.regOrder[i]
		if ctn, ok := o.items[name]; ok {
			o.log.Debugf("closing container [%s]...", name)
			err := ctn.Close(ctx)
			if err != nil {
				// Мягкое гашение: только логируем ошибку, не прерывая деструкцию остальных контейнеров кластера
				o.log.Errorf("failed to close container [%s]: %v", name, err)
				closeErrs = append(closeErrs, err)
			}
		}
	}

	// Агрегируем пачку ошибок в единую цепочку без потери исходных контекстов
	err := errors.Join(closeErrs...)
	if err != nil {
		return errs.NewContainerError("orchestrator", "close containers failed", err)
	}

	return nil
}

// Register атомарно добавляет новый контейнер в реестр под эксклюзивным write-локом с пре-валидацией.
func (o *BaseOrchestrator) Register(container Container) error {
	o.log.Debug("Register start")
	defer o.log.Debug("Register finish")

	if err := o.validateContainer("BaseOrchestrator.Register", container); err != nil {
		return err
	}
	if o.HasContainer(container.GetName()) {
		return errs.NewContainerError("orchestrator", fmt.Sprintf("container [%s] already registered", container.GetName()), nil)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.items[container.GetName()] = container
	o.regOrder = append(o.regOrder, container.GetName())

	return nil
}

// Unregister принудительно удаляет контейнер из мапы ресурсов и вырезает его имя из слайса порядка вызовов.
func (o *BaseOrchestrator) Unregister(name string) error {
	o.log.Debug("Unregister start")
	defer o.log.Debug("Unregister finish")

	if err := o.validateName("BaseOrchestrator.Unregister", name); err != nil {
		return err
	}
	if !o.HasContainer(name) {
		return errs.NewContainerError("orchestrator", fmt.Sprintf("container [%s] not registered", name), nil)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.items, name)
	// Go 1.21+ оптимизация: эффективная lock-free фильтрация слайса без выделения нового массива
	o.regOrder = slices.DeleteFunc(o.regOrder, func(item string) bool {
		return item == name
	})

	return nil
}

// GetContainer извлекает контейнер по имени под быстрой read-блокировкой.
func (o *BaseOrchestrator) GetContainer(name string) (Container, error) {
	o.log.Debug("GetContainer start")
	defer o.log.Debug("GetContainer finish")

	o.mu.RLock()
	defer o.mu.RUnlock()

	res, ok := o.items[name]
	if !ok {
		return nil, errs.NewContainerError("orchestrator", fmt.Sprintf("container [%s] not found", name), nil)
	}

	return res, nil
}

// HasContainer проверяет фактическое существование контейнера в реестре.
func (o *BaseOrchestrator) HasContainer(name string) bool {
	o.log.Debug("HasContainer start")
	defer o.log.Debug("HasContainer finish")

	o.mu.RLock()
	defer o.mu.RUnlock()

	_, ok := o.items[name]
	return ok
}

// AllContainers возвращает хронологически отсортированный срез всех живых контейнеров приложения.
func (o *BaseOrchestrator) AllContainers() []Container {
	o.log.Debug("AllContainers start")
	defer o.log.Debug("AllContainers finish")

	o.mu.RLock()
	defer o.mu.RUnlock()

	res := make([]Container, 0, len(o.items))
	for _, name := range o.regOrder {
		res = append(res, o.items[name])
	}

	return res
}

// GetRunners сканирует все зарегистрированные контейнеры, извлекает их инстансы и динамически
// отбирает те компоненты, которые удовлетворяют интерфейсу Runner, подготавливая их к запуску в main.go.
func (o *BaseOrchestrator) GetRunners() ([]Runner, error) {
	o.log.Debug("GetRunners start")
	defer o.log.Debug("GetRunners finish")

	o.mu.RLock()
	defer o.mu.RUnlock()

	var res []Runner
	for _, name := range o.regOrder {
		ctn := o.items[name]
		for _, instName := range ctn.AllNames() {
			inst, err := ctn.GetInstance(instName)
			if err != nil {
				return nil, errs.NewContainerError("orchestrator", fmt.Sprintf("get instance [%s] failed", instName), err)
			}
			// Динамическое приведение типов (Type Assertion): вычленяем исполняемые горутины-раннеры
			if r, ok := inst.(Runner); ok {
				res = append(res, r)
			}
		}
	}

	return res, nil
}

// validateName — внутренний защитный пре-валидатор строкового идентификатора.
func (o *BaseOrchestrator) validateName(op, name string) error {
	if name == "" {
		return errs.NewContainerValidateError("orchestrator", op, "name is empty", nil)
	}

	return nil
}

// validateContainer — внутренний защитный пре-валидатор структуры интерфейса.
func (o *BaseOrchestrator) validateContainer(op string, container Container) error {
	if utils.IsNil(container) {
		return errs.NewContainerValidateError("orchestrator", op, "container is nil", nil)
	}

	return nil
}
