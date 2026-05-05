/*
Package container является центральным узлом управления зависимостями (DI) приложения.

Пакет реализует логику сборки графа объектов и управляет их жизненным циклом
с использованием оркестратора (pkg/container/Orchestrator).

Архитектурные принципы:

 1. Инкапсуляция сборки: Здесь описывается, КАК создаются объекты, но не их логика.
 2. Ленивая инициализация: Благодаря BaseLazyContainer ресурсы (например, БД)
    инициализируются только при первом обращении.
 3. Слоистая структура: Пакет разделен на логические файлы для упрощения навигации:
    Рекомендуемый набор контейнеров и последовательность инициализации:
    - app.go: Конфигурация, логгер
    - tools.go: Утилиты, инструментарий (utils and helpers).
    - db.go или postgres.go/oracle.go/mssql.go/etc.: Соединение с БД, миграция данных
    - rest/gRPC clients: Клиенты rest/gRPC
    - repository.go: Уровень доступа к данным (Repositories).
    - use_case.go: Уровень бизнес-логики (Domain Services).
    - facade.go: Уровень транспорта приложения (конечная точка)
    - http_service.go: HTTP router and middleware
    - grpc_service.go: gRPC services
    - worker.go: workers Runners
    - transport.go: Сетевые адаптеры (HTTP/gRPC Runners) и gRPC interceptors.

Использование:
Контейнеры регистрируются в Оркестраторе в app.go, после чего управление
передается объекту BaseApplication для инициализации и запуска.

Пример регистрации:

	orch := container.NewBaseOrchestrator(log)
	infraCont := infra.NewContainer(orch, conf, log)
	orch.Register(infraCont)
*/
package container
