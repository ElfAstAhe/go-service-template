package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseApplication реализует интерфейс Application, координируя центральный четырехфазный рантайм-жизненный цикл микросервиса.
//
// Инкапсулирует глобальный контекст отмены (Cancellation Context), осуществляет оркестрацию запуска асинхронных раннеров (Runners),
// управляет перехватом системных сигналов ОС и гарантирует строго контролируемый по времени Graceful Shutdown.
type BaseApplication struct {
	conf         *config.AppConfig      // Ссылка на центральные таймауты и параметры среды (dev/prod/test)
	orchestrator container.Orchestrator // Ссылка на глобальный IoC/DI оркестратор графа зависимостей компонентов
	logger       logger.Logger          // Структурированный логгер верхнего уровня для паспортизации шагов рантайма
	ctx          context.Context        // Корневой контекст приложения, контролирующий длительность жизни асинхронных горутин
	cancel       context.CancelFunc     // Функция-триггер для атомарной отмены контекста и запуска остановки подов
	wg           sync.WaitGroup         // WaitGroup контроля завершения системных горутин верхнего уровня
	ready        *atomic.Bool           // Атомарный флаг готовности (Readiness) для интеграции с Kubernetes Health Probes
}

var _ Application = (*BaseApplication)(nil)

// NewBaseApplication — фабричный конструктор оркестратора жизненного цикла приложения.
func NewBaseApplication(opts ...Option) *BaseApplication {
	// Инициализируем изолированный корневой контекст выполнения
	ctx, cancel := context.WithCancel(context.Background())

	res := &BaseApplication{
		conf:   config.NewDefaultAppConfig(), // Загружаем константные дефолты таймаутов рантайма
		ctx:    ctx,
		cancel: cancel,
		ready:  new(atomic.Bool),
	}

	// Применяем мутаторы конфигурации Fluent API
	for _, opt := range opts {
		opt(res)
	}
	res.ready.Store(false)

	return res
}

// Init выполняет первую фазу холодного старта (Bootstrap phase), каскадно инициализируя граф DI-контейнеров.
func (app *BaseApplication) Init() error {
	if utils.IsNil(app.orchestrator) {
		return errs.NewCommonError("orchestrator is nil", nil)
	}

	// Жестко лимитируем время сборки и выделения ресурсов контейнерами (Fail-Fast паттерн)
	initCtx, cancel := context.WithTimeout(app.ctx, app.conf.InitTimeout)
	defer cancel()

	if err := app.orchestrator.Init(initCtx); err != nil {
		return errs.NewCommonError("orchestrator init failed", err)
	}

	return nil
}

// Run переводит микросервис в активную фазу исполнения, блокируя вызывающий поток до отмены контекста или сигналов ОС.
func (app *BaseApplication) Run() error {
	// 1. Асинхронно запускаем веер всех извлеченных раннеров (HTTP/gRPC/Workers) в горутинах
	if err := app.Start(); err != nil {
		return errs.NewCommonError("failed to start runners", err)
	}

	// 2. Включаем фоновый слушатель POSIX сигналов ОС (SIGTERM, SIGINT) в отдельном потоке
	app.wg.Add(1)
	go app.GracefulShutdown()

	// Выставляем атомарный флаг готовности к приему трафика кластера
	app.ready.Store(true)

	app.logger.Info("application is running and waiting for app context cancel")
	// 3. Блокируем главный поток main.go, ожидая сигнала отмены корневого контекста
	<-app.ctx.Done()

	// Снимаем статус готовности: K8s перестает направлять трафик на данный под
	app.ready.Store(false)

	// 4. Запускаем фазу Graceful Shutdown для активных раннеров
	app.logger.Info("application is shutting down")
	err := app.Stop()

	// Ожидаем физического завершения всех зарегистрированных горутин верхнего уровня
	app.logger.Info("application is waiting for runners stopped")
	app.WaitForStop()

	return err
}

// Start извлекает из графа DI все исполняемые компоненты и запускает их параллельно с защитой от "тихих падений".
func (app *BaseApplication) Start() error {
	runners, err := app.orchestrator.GetRunners()
	if err != nil {
		return err
	}

	for _, r := range runners {
		app.wg.Add(1)
		go func(runner container.Runner) {
			defer app.wg.Done()
			app.logger.Infof("runner [%s] starting", runner.GetName())

			// Запускаем бесконечный цикл обработки раннера с пробросом корневого контекста
			if err := runner.Start(app.ctx); err != nil {
				app.logger.Errorf("runner [%s] failed: %v", runner.GetName(), err)
				// 🛡️ Защитный барьер: при падении любого раннера гасим все приложение для предотвращения зомби-процессов
				app.cancel()
			}
		}(r)
	}

	return nil
}

// Stop выполняет фазу контролируемой мягкой остановки раннеров, блокируя их новые операции ввода-вывода.
func (app *BaseApplication) Stop() error {
	app.logger.Info("stopping active runners (graceful shutdown phase)...")

	// Дублируем отмену контекста для каскадного уведомления всех дочерних структур
	app.cancel()

	runners, err := app.orchestrator.GetRunners()
	if err != nil {
		return err
	}

	// Создаем выделенный context.Background() с таймаутом останова для изоляции от уже отмененного app.ctx
	stopCtx, stopCancel := context.WithTimeout(context.Background(), app.conf.StopTimeout)
	defer stopCancel()

	var (
		stopWg sync.WaitGroup
		mu     sync.Mutex
		// ToDo: переделать на ConcurrentList
		stopErrs []error
	)

	// Параллельно веерно (Fan-Out) гасим каждый раннер в рамках выделенного лимита времени
	for _, r := range runners {
		stopWg.Add(1)
		go func(runner container.Runner) {
			defer stopWg.Done()
			if err := runner.Stop(stopCtx); err != nil {
				mu.Lock()
				stopErrs = append(stopErrs, err)
				mu.Unlock()
			}
		}(r)
	}

	stopWg.Wait()
	return errors.Join(stopErrs...) // Агрегируем пачку ошибок закрытия интерфейсов в единую цепочку Go 1.22+
}

// Close осуществляет финальный такт тотального уничтожения и деаллокации пулов памяти/сокетов СУБД.
func (app *BaseApplication) Close() error {
	app.logger.Info("closing application resources (containers)...")

	// Создаем выделенный контекст для фазы деструкции ресурсов СУБД/Кафки
	closeCtx, cancel := context.WithTimeout(context.Background(), app.conf.CloseTimeout)
	defer cancel()

	// Делегируем очистку IoC-оркестратору, который пройдет по контейнерам в строгом LIFO-порядке
	if err := app.orchestrator.Close(closeCtx); err != nil {
		return errs.NewCommonError("orchestrator close failed", err)
	}

	app.logger.Info("application resources closed successfully")
	return nil
}

// GracefulShutdown метод мягкого закрытия приложения,
// слушает сигналы ос или контекст приложения, в случае сигнала OS отменяет контекст приложения,
// по приходу сигнала или отмены контеста приложения метод завершает свою работу
//
// Пример использования:
//
//	go app.GracefulShutdown()
func (app *BaseApplication) GracefulShutdown() {
	defer app.wg.Done() // Сигнализируем о завершении работы системного перехватчика в общую WaitGroup

	log := app.logger.GetLogger("BaseApplication.GracefulShutdown")
	log.Debugf("Graceful shutdown goroutine started")

	// Аллоцируем буферизованный канал для приема системных прерываний (размер 1 обязателен по спецификации пакета signal)
	osSigChan := make(chan os.Signal, 1)

	// Пул отслеживаемых POSIX сигналов ядра операционной системы Linux/Unix
	osSignals := []os.Signal{
		syscall.SIGINT,  // Сигнал прерывания с клавиатуры (Ctrl+C)
		syscall.SIGTERM, // Основной сигнал остановки контейнера оркестратором Kubernetes / Docker
		syscall.SIGQUIT, // Сигнал экстренного завершения с дампом памяти (Ctrl+\)
	}

	// Регистрируем наш канал в системном ядре Go для перехвата указанных сигналов
	signal.Notify(osSigChan, osSignals...)
	// 🛡️ Защитный барьер: дефер гарантированно вычищает подписку в ядре Go, предотвращая Memory Leaks
	defer signal.Stop(osSigChan)

	log.Debug("listening signals")
	for _, sig := range osSignals {
		log.Debugf("os signal: [%s]", sig.String())
	}

	// Двунаправленный неблокирующий барьер ожидания наступления события останова
	select {
	case osSig := <-osSigChan:
		// Сценарий 1: Прилетел сигнал от Kubernetes (SIGTERM). Начинаем процедуру веерной отмены.
		log.Debugf("received os signal [%s], cancel main app context", osSig.String())
		app.cancel() // Триггерим каскадную отмену контекстов всех раннеров и воркеров
	case <-app.ctx.Done():
		// Сценарий 2: Корневой контекст приложения уже был отменен изнутри (например, из-за падения воркера)
		log.Debug("main app context has been canceled")
	}
}

// WaitForStop блокирует вызывающий поток до полного завершения всех зарегистрированных в WaitGroup горутин фреймворка.
func (app *BaseApplication) WaitForStop() {
	app.wg.Wait()
}

// GetWaitGroup возвращает ссылку на внутреннюю структуру sync.WaitGroup координатора.
func (app *BaseApplication) GetWaitGroup() *sync.WaitGroup {
	return &app.wg
}

// GetLogger возвращает инстанс центрального структурированного логгера приложения.
func (app *BaseApplication) GetLogger() logger.Logger {
	return app.logger
}

// GetOrchestrator возвращает ссылку на глобальный IoC/DI оркестратор графа зависимостей.
func (app *BaseApplication) GetOrchestrator() container.Orchestrator {
	return app.orchestrator
}

// GetContext возвращает ссылку на корневой контекст времени жизни приложения.
func (app *BaseApplication) GetContext() context.Context {
	return app.ctx
}

// GetCancel возвращает функцию-триггер экстренной отмены корневого контекста рантайма.
func (app *BaseApplication) GetCancel() context.CancelFunc {
	return app.cancel
}

// GetConfig возвращает ссылку на конфигурационный паспорт параметров ядра приложения.
func (app *BaseApplication) GetConfig() *config.AppConfig {
	return app.conf
}

// IsReady возвращает атомарный статус готовности приложения (Readiness статус) для K8s Health Probes.
func (app *BaseApplication) IsReady() bool {
	return app.ready.Load()
}
