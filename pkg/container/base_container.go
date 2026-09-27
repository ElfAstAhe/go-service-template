package container

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseContainer — базовая потокобезопасная реализация реестра инстансов (Dependency Injection Container).
//
// Инкапсулирует в себе плоскую карту нетипизированных объектов any, защищенную мьютексом (sync.RWMutex).
// Отвечает за параллельный, неблокирующий запуск деструкторов компонентов (Simple/Context Closers)
// строго в рамках отведенных лимитов времени контекста деплоя (Graceful Shutdown).
type BaseContainer struct {
	name         string         // Уникальное текстовое имя контейнера (например, "PgContainer")
	mu           sync.RWMutex   // RWMutex для безопасной изоляции параллельных операций чтения/записи горутин
	instances    map[string]any // Внутреннее хранилище живых, инициализированных объектов-зависимостей
	orchestrator Orchestrator   // Ссылка на центральный IoC-оркестратор для интеграции в глобальный граф
	logger       logger.Logger  // Изолированный структурированный логгер контейнера
}

// NewBaseContainer — фабричный конструктор базового контейнера зависимостей.
func NewBaseContainer(
	opts ...Option,
) *BaseContainer {
	options := &Options{}

	// Вычисляем конфигурационные мутаторы Fluent API
	for _, o := range opts {
		o(options)
	}

	return &BaseContainer{
		name:         options.Name,
		instances:    make(map[string]any),
		orchestrator: options.Orchestrator,
		logger:       options.Logger.GetLogger("BaseContainer"),
	}
}

// Гарантируем соответствие базовому интерфейсу Container на этапе компиляции
var _ Container = (*BaseContainer)(nil)

// GetName возвращает строковый идентификатор текущего контейнера.
func (bc *BaseContainer) GetName() string {
	return bc.name
}

// Init — метод инициализации ресурсов контейнера.
// По умолчанию возвращает ошибку NotImplementedError, делегируя расширение специализированным контейнерам.
func (bc *BaseContainer) Init(initCtx context.Context) error {
	return errs.NewNotImplementedError(nil)
}

// Close выполняет параллельную веерную очистку (Fan-Out Cleanup) всех закрываемых ресурсов.
// Гарантирует атомарное тушение сокетов и пулов строго до истечения таймаута контекста closeCtx.
//
//goland:noinspection DuplicatedCode
func (bc *BaseContainer) Close(closeCtx context.Context) error {
	bc.logger.Debugf("container %s close: started", bc.GetName())
	defer bc.logger.Debugf("container %s close: finished", bc.GetName())

	var wg sync.WaitGroup

	// 1. Быстро копируем ссылки на инстансы под RLock для минимизации времени блокировки
	bc.mu.RLock()
	toClosing := make([]*closeInstance, 0, len(bc.instances))
	for name, instance := range bc.instances {
		toClosing = append(toClosing, newCloseInstance(name, instance))
	}
	bc.mu.RUnlock()

	bc.logger.Debugf("container %s close: got %d instances to close", bc.GetName(), len(toClosing))

	// 2. Полностью очищаем мапу под Lock, так как контейнер уничтожается рантаймом
	bc.mu.Lock()
	bc.instances = make(map[string]any)
	bc.mu.Unlock()

	closeChan := make(chan struct{})
	closeErrs := utils.NewConcurrentList[error]() // Используем потокобезопасный список для агрегации ошибок

	// 3. Запускаем параллельный цикл деструкции компонентов
	for _, toClose := range toClosing {
		if inst, ok := toClose.Instance.(SimpleCloser); ok {
			wg.Add(1)
			go func(name string, closer SimpleCloser) {
				bc.logger.Debugf("container %s close: simple closer for instance %s start", bc.GetName(), name)
				defer bc.logger.Debugf("container %s close: simple closer for instance %s finish", bc.GetName(), name)
				defer wg.Done()

				if err := closer.Close(); err != nil {
					closeErrs.Append(err)
					bc.logger.Debugf("container %s close: simple closer for instance %s failed: %v", bc.GetName(), name, err)
				} else {
					bc.logger.Debugf("container %s close: simple closer for instance %s done", bc.GetName(), name)
				}
			}(toClose.Name, inst)
		} else if inst, ok := toClose.Instance.(ContextCloser); ok {
			wg.Add(1)
			go func(name string, closer ContextCloser) {
				bc.logger.Debugf("container %s close: context closer for instance %s start", bc.GetName(), name)
				defer bc.logger.Debugf("container %s close: context closer for instance %s finish", bc.GetName(), name)
				defer wg.Done()

				if err := closer.Close(closeCtx); err != nil {
					closeErrs.Append(err)
					bc.logger.Debugf("container %s close: context closer for instance %s failed: %v", bc.GetName(), name, err)
				} else {
					bc.logger.Debugf("container %s close: context closer for instance %s done", bc.GetName(), name)
				}
			}(toClose.Name, inst)
		} else {
			bc.logger.Debugf("container %s close: no closer methods for instance %s", bc.GetName(), toClose.Name)
		}
	}

	// Сигнальная фоновая горутина ожидания завершения пула деструкторов
	go func() {
		defer close(closeChan)
		wg.Wait()
	}()

	// 4. Селектор контроля таймаута мягкой остановки (Graceful Shutdown Shield)
	select {
	case <-closeChan:
		if closeErrs.Len() > 0 {
			return errs.NewContainerError(bc.GetName(), "container close: close fails", errors.Join(closeErrs.Snapshot()...))
		}
		return nil
	case <-closeCtx.Done():
		// Защита от зависания пода: выходим по жесткому таймауту конфигурации пода Kubernetes
		return errs.NewContainerError(bc.GetName(), "container close: close timeout limit reached", nil)
	}
}

// RegisterInstance атомарно регистрирует готовый инстанс зависимости во внутренней мапе под write-локом.
func (bc *BaseContainer) RegisterInstance(name string, instance any) error {
	bc.logger.Debugf("container %s register instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s register instance: finished", bc.GetName())

	if err := bc.Validate("BaseContainer.RegisterInstance", name); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Предохранитель: исключает перезапись и коллизии имен в рамках одного модуля
	if _, ok := bc.instances[name]; ok {
		return errs.NewContainerError(bc.GetName(), fmt.Sprintf("instance %s already exists", name), nil)
	}

	bc.instances[name] = instance
	return nil
}

// UnregisterInstance атомарно вырезает инстанс из реестра памяти контейнера.
func (bc *BaseContainer) UnregisterInstance(name string) error {
	bc.logger.Debugf("container %s unregister instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s unregister instance: finished", bc.GetName())

	if err := bc.Validate("BaseContainer.UnregisterInstance", name); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	delete(bc.instances, name)
	return nil
}

// GetInstance извлекает нетипизированную зависимость под быстрой read-блокировкой.
func (bc *BaseContainer) GetInstance(name string) (any, error) {
	bc.logger.Debugf("container %s get instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s get instance: finished", bc.GetName())

	if err := bc.Validate("BaseContainer.GetInstance", name); err != nil {
		return nil, err
	}

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	res, ok := bc.instances[name]
	if !ok {
		return nil, errs.NewContainerNotFoundError(fmt.Sprintf("container [%s] instance [%s] is not registered ", bc.GetName(), name), nil)
	}

	return res, nil
}

// AllNames собирает изолированную копию всех строковых идентификаторов, удерживаемых в текущем контейнере.
func (bc *BaseContainer) AllNames() []string {
	bc.logger.Debugf("container %s all names: started", bc.GetName())
	defer bc.logger.Debugf("container %s all names: finished", bc.GetName())

	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if len(bc.instances) == 0 {
		return nil
	}

	res := make([]string, 0, len(bc.instances))
	for key := range bc.instances {
		res = append(res, key)
	}

	return res
}

// Validate выполняет оборонительную проверку (Guard Clause) входящих аргументов контейнера.
// Пресекает попытки оперирования пустыми идентификаторами зависимостей на проде.
func (bc *BaseContainer) Validate(op string, name string) error {
	bc.logger.Debugf("container %s validate: started", bc.GetName())
	defer bc.logger.Debugf("container %s validate: finished", bc.GetName())

	if name == "" {
		return errs.NewContainerValidateError(bc.GetName(), op, "name is empty", nil)
	}

	return nil
}

// IsRegistered проверяет под быстрой read-блокировкой, зарегистрирован ли компонент с указанным именем.
func (bc *BaseContainer) IsRegistered(name string) bool {
	bc.logger.Debugf("container %s is registered: started", bc.GetName())
	defer bc.logger.Debugf("container %s is registered: finished", bc.GetName())

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	_, ok := bc.instances[name]

	return ok
}

// HasInstance проверяет фактическое наличие живого инициализированного объекта в оперативной памяти.
// Для плоского не-ленивого контейнера логика полностью идентична методу IsRegistered.
func (bc *BaseContainer) HasInstance(name string) bool {
	bc.logger.Debugf("container %s has instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s has instance: finished", bc.GetName())

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	_, ok := bc.instances[name]

	return ok
}

// GetOrchestrator возвращает прямую ссылку на инжектированный центральный IoC-оркестратор графа зависимостей.
func (bc *BaseContainer) GetOrchestrator() Orchestrator {
	return bc.orchestrator
}
