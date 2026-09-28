package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// ServerProvider определяет сигнатуру фабричной функции-провайдера для сборки экземпляра веб-сервера.
type ServerProvider func(router Router, conf *config.HTTPConfig) (*http.Server, error)

// ServerLauncher определяет сигнатуру функции запуска слушателя сетевых сокетов HTTP-сервера.
type ServerLauncher func(server *http.Server, conf *config.HTTPConfig) error

// Runner реализует интерфейс container.Runner, выступая в роли низкоуровневого оркестратора
// и диспетчера жизненного цикла HTTP/REST веб-сервера приложения.
//
// Инкапсулирует под своим крылом сетевые конфигурации пулов сокетов, инжектированные роутеры,
// кастомные лаунчеры и берет на себя атомарную координацию фазы мягкого гашения (Graceful Shutdown).
type Runner struct {
	name           string             // Уникальное текстовое имя раннера для построения контекстов логирования
	router         Router             // Инжектированный сетевой роутер, предоставляющий http.Handler
	server         *http.Server       // Живой низкоуровневый экземпляр веб-сервера стандартной библиотеки Go
	conf           *config.HTTPConfig // Ссылка на конфигурационный паспорт параметров сетевого HTTP-слоя
	running        *atomic.Bool       // Атомарный флаг активности, защищающий от двойного запуска/останова
	serverProvider ServerProvider     // Инжектированная стратегия сборки параметров сервера
	serverLauncher ServerLauncher     // Инжектированная стратегия запуска сетевого слушателя
	log            logger.Logger      // Изолированный структурированный логгер компонента
}

// Гарантируем строгое соответствие контракту Runner на этапе компиляции
var _ container.Runner = (*Runner)(nil)

// NewRunner — фабричный конструктор HTTP-раннера фреймворка.
// Применяет пачку функциональных опций Fluent API и производит обязательную оборонительную валидацию (Guard Clauses).
func NewRunner(
	opts ...Option,
) (*Runner, error) {
	res := &Runner{
		name:    "default http",
		conf:    config.NewDefaultHTTPConfig(), // Наполняем базовыми таймаутами по умолчанию
		running: new(atomic.Bool),
	}
	res.running.Store(false)

	// Вычисляем входящие мутаторы конфигурации Fluent API
	for _, option := range opts {
		option(res)
	}

	// Оборонительная проверка (Guard Clauses): пресекаем запуск без критических зависимостей
	if utils.IsNil(res.router) {
		return nil, errs.NewTlCommonError("NewRunner", "router not applied", nil)
	}
	if utils.IsNil(res.log) {
		return nil, errs.NewTlCommonError("NewRunner", "logger not applied", nil)
	}

	// Ленивая подстановка системных дефолтных провайдеров при отсутствии кастомизации
	if utils.IsNil(res.serverProvider) {
		res.serverProvider = res.defaultServerProvider
	}
	if utils.IsNil(res.serverLauncher) {
		res.serverLauncher = res.defaultServerLauncher
	}

	return res, nil
}

// Start осуществляет сборку и непосредственный запуск HTTP сетевого слушателя сокетов.
// ВНИМАНИЕ! Данный метод является блокирующим (блокирует горутину вызова бесконечным Event Loop сервера).
// Корректно обрабатывает каноничную ошибку закрытия сокета http.ErrServerClosed.
func (r *Runner) Start(ctx context.Context) error {
	r.log.Debugf("Runner.Start %s start", r.GetName())
	defer r.log.Debugf("Runner.Start %s finish", r.GetName())

	// Атомарный CAS-барьер защиты от повторных или гоночных вызовов старта
	if !r.running.CompareAndSwap(false, true) {
		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s already started", r.GetName()), nil)
	}
	var err error

	// Шаг 1: Потокобезопасная аллокация сервера через инжектированный провайдер
	r.server, err = r.serverProvider(r.router, r.conf)
	if err != nil {
		r.running.Store(false) // Откатываем флаг при сбое сборки
		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s create http server failed", r.GetName()), err)
	}

	// Шаг 2: Запуск бесконечного цикла прослушивания входящих сетевых Tls/Plain соединений
	err = r.serverLauncher(r.server, r.conf)
	// Игнорируем штатную ошибку ErrServerClosed, возникающую при плановой остановке сокета методом Shutdown
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		r.running.Store(false)
		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s http server listen failed", r.GetName()), err)
	}

	return nil
}

// Stop инициирует процедуру каноничного мягкого гашения (Graceful Shutdown) веб-сервера.
// Вырезает ссылку на активный сервер, создает изолированный контекст с таймаутом ShutdownTimeout
// и плавно дожидается завершения обслуживания всех текущих сетевых HTTP-сессий клиентов.
func (r *Runner) Stop(stopCtx context.Context) error {
	r.log.Debugf("Runner.Stop %s start", r.GetName())
	defer r.log.Debugf("Runner.Stop %s finish", r.GetName())

	// Атомарный CAS-барьер защиты от вызова останова на неработающем компоненте
	if !r.running.CompareAndSwap(true, false) {
		return errs.NewTlCommonError("Runner.Stop", fmt.Sprintf("runner %s not running", r.GetName()), nil)
	}

	var srv *http.Server
	// Атомарно вырезаем ссылку на рабочий сервер во избежание Race Conditions с методом Start
	srv, r.server = r.server, nil

	// Создаем изолированный дочерний контекст с таймаутом мягкого гашения
	shutdownCtx, shutdownCancel := context.WithTimeout(stopCtx, r.conf.ShutdownTimeout)
	defer shutdownCancel()

	// Вызываем нативный системный Graceful Shutdown стандартной библиотеки Go
	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Предупреждение: сервер не успел доварить сессии в жесткий тайм-лимит, сокеты закрываются силой
			r.log.Warn("http server shutdown timed out (force close)")
		} else {
			r.log.Errorf("http server shutdown with error [%v]", err)
		}

		return errs.NewTlCommonError("Runner.Stop", "http server shutdown with error", err)
	}
	r.log.Debug("http server shutdown gracefully complete")

	return nil
}

// GetName возвращает уникальное строковое наименование текущего экземпляра раннера.
func (r *Runner) GetName() string {
	return r.name
}

// IsRunning возвращает текущий атомарный статус активности HTTP сокета (true — работает и принимает трафик).
func (r *Runner) IsRunning() bool {
	return r.running.Load()
}

// defaultServerProvider — внутренний (неэкспортируемый) базовый сборщик параметров http.Server.
func (r *Runner) defaultServerProvider(router Router, conf *config.HTTPConfig) (*http.Server, error) {
	r.log.Debugf("Runner.defaultServerProvider %s start", r.GetName())
	defer r.log.Debugf("Runner.defaultServerProvider %s finish", r.GetName())

	return &http.Server{
		Addr:         conf.Address,
		Handler:      router.GetRouter(),
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
	}, nil
}

// defaultServerLauncher — внутренний (неэкспортируемый) базовый исполнитель ListenAndServe сетевого сокета.
func (r *Runner) defaultServerLauncher(server *http.Server, conf *config.HTTPConfig) error {
	r.log.Debugf("Runner.defaultServerLauncher %s start", r.GetName())
	defer r.log.Debugf("Runner.defaultServerLauncher %s finish", r.GetName())

	// Маршрутизируем запуск на основе флага безопасности (HTTPS mTLS контур / Plain HTTP)
	if conf.Secure {
		r.log.Infof("Runner.defaultServerLauncher %s http secure listen %s", r.GetName(), conf.Address)
		return server.ListenAndServeTLS(conf.CertificatePath, conf.PrivateKeyPath)
	}

	r.log.Infof("Runner.defaultServerLauncher %s http nonsecure listen %s", r.GetName(), conf.Address)
	return server.ListenAndServe()
}
