# Шаблон сервиса/микро-сервиса, уровень go-middle

## 🏗 Структура проекта (Project Layout)

Проект придерживается принципов **Clean Architecture** и **Standard Go Project Layout**.

```text
.
├── api/                # Контракты: OpenAPI/Swagger спецификации/конфигурации, Proto-файлы
│   ├── proto/          # Proto файлы сервиса
│   └── rest/           # Конфигурация генератора клиента, etc. (спорно)
├── bin/                # Артефакты сборки
├── cmd/                # Приложения
│   └── app/            # Точка входа: инициализация DI и запуск приложения
├── deployments/        # Конфигурация инфраструктуры: Dockerfile, docker-compose, k8s
├── docs/               # Сгенерированная документация (Swagger UI)
├── internal/           # Приватный код приложения (бизнес-логика)
│   ├── app/            # Оркестрация: инициализация всех слоев, Graceful Shutdown
│   ├── config/         # Конфигурация: загрузка YAML/ENV/Flags
│   ├── domain/         # Сердце системы: доменные модели (Entities) и базовые интерфейсы
│   │   ├── errs/       # BLL ошибки
│   │   └── mocks/      # Mock файлы 
│   ├── facade/         # Фасад: внешняя граница приложения
│   │   ├── dto/        # Фасад: dto
│   │   └── mapper/     # Фасад: mappers
│   ├── usecase/        # Бизнес-логика: реализация сценариев использования
│   ├── repository/     # DAL: реализация работы с БД, кешем и внешними API
│   │   └── postgres/   # DAL: реализация для postgres
│   └── transport/      # Транспортный слой (Внешние интерфейсы)
│       ├── errs/       # Ошибки уровня транспорта
│       ├── rest/       # HTTP: router, handlers, middleware, DTO, mappers
│       └── grpc/       # gRPC: Реализация сервисов и интерцепторы
├── migrations/         # Миграции базы данных (in-code/sql)
│   └── example-service # Миграции БД сервиса exanple-service
├── pkg/                # Публичные библиотеки (Logger, Auth, Errors, Utils)
│   ├── api/            # Клиенты, автогенерация
│   ├── auth/           # Аутентификация
│   ├── config/         # Конфигурирование
│   ├── db/             # Абстракция БД
│   ├── domain/         # Абстракция domain model
│   ├── errs/           # Общие ошибки
│   ├── helper/         # helpers
│   ├── infra/          # Инфраструктура
│   │   ├── metrics/    # Метрики
│   │   └── telemetry/  # Телеметрия (open tracing)
│   ├── logger/         # Логирование
│   ├── migration/      # миграция данных
│   ├── repository/     # реализация CRUD репозитория
│   ├── transport/      # транспорт
│   │   └── middleware/ # middleware
│   └── utils/          # утилиты
├── scripts/            # Вспомогательные скрипты для разработки и CI/CD
├── .gitignore          # Список файлов к пропуску для git 
├── .mockery.yml        # Конфигурация генератора mock файлов 
├── Makefile            # Команды автоматизации (build, run, test, swag, migrate)
├── go.mod              # Зависимости
├── go.sum              # Контрольные суммы
└── readme.md           # Документация проекта
```
## Описание слоев
1. `Domain`: Не зависит ни от чего. Содержит только структуры данных и интерфейсы репозиториев/сервисов.
2. `Usecase`: Содержит бизнес-логику. Зависит только от Domain.
3. `Repository (Adapter)`: Реализация интерфейсов из Domain. Здесь живет SQL, работа с Redis или внешними клиентами.
4. `Transport (Delivery)`: Входные точки. Превращают внешние запросы (JSON, Protobuf) в данные, понятные слою Usecase.
5. `App`: "Склеивает" все слои воедино (Dependency Injection).
6. `Pkg`: Набор утилит, которые можно без изменений перенести в любой другой проект.

## В example-service реализовано:
* transaction manager for use cases and context transaction support
* CRUD generic repository pattern with context transaction support
* repository metrics and tracing via decorator pattern
* use cases with single responsibility
* CRUD facade
* http with chi router and pipeline setup(tracing, metrics, logging, compress and decompress)
* gRPC with standard google library and pipeline setup (tracing, metrics)
* application configuration with viper library
* and application skeleton itself :-)

## Цепочка вызовов
### `transport -> facade -> use case -> repository -> lib/helper` и обратно