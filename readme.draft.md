# Шаблон сервиса/микро-сервиса, уровень go-middle

## 🏗 Структура проекта (Project Layout)

Проект придерживается принципов **Clean Architecture** и **Standard Go Project Layout**.

```text
.
├── api/                    # Контракты: OpenAPI/Swagger спецификации/конфигурации, Proto-файлы
│   ├── proto/              # Proto файлы сервиса
│   └── rest/               # Конфигурация генератора клиента, тестовые http запросы/коллекция
├── bin/                    # Артефакты сборки
├── cmd/                    # Приложения
│   ├── gen-tz/             # Утилита кодогенерации констант таймзон
│   └── example-service/    # Точка входа примера сервиса
├── configs/                # Документация по настройке дополнительных систем (brokers,cache,etc)
├── deployments/            # Конфигурация инфраструктуры: Dockerfile, docker-compose, k8s
├── docs/                   # Сгенерированная документация (Swagger UI)
├── internal/               # Приватный код сервиса
│   ├── app/                # Оркестрация: инициализация всех слоев, Graceful Shutdown
│   │   └── container/      # DI контейнеры 
│   ├── config/             # Конфигурация: загрузка YAML/ENV/Flags
│   ├── domain/             # BLL-dom: модели (Entities), интерфейсы (repository/services/etx)
│   │   └── mocks/          # Mock файлы 
│   ├── facade/             # Фасад: внешняя граница приложения (по сути внешний интерфейс)
│   │   ├── dto/            # Фасад: dto
│   │   └── mapper/         # Фасад: mappers
│   ├── usecase/            # BLL-uc: Бизнес-логика: реализация сценариев использования
│   │   └── trace/          # BLL: телеметрия
│   ├── repository/         # DAL: реализация работы с БД, кешем и внешними API
│   │   ├── metrics/        # DAL: метрики
│   │   ├── trace/          # DAL: телеметрия
│   │   └── postgres/       # DAL: реализация для postgres
│   └── transport/          # TL: Транспортный слой
│       ├── rest/           # TL: rest/HTTP: router, handlers, middleware, DTO, mappers
│       └── grpc/           # gRPC: Реализация сервисов и интерцепторы
├── migrations/             # Миграции базы данных (in-code/sql)
│   └── example-service     # Миграции БД сервиса exanple-service
├── pkg/                    # Публичные библиотеки (Logger, Auth, Errors, Utils)
│   ├── api/                # Клиенты, автогенерация
│   ├── app/                # Приложение
│   ├── auth/               # Аутентификация
│   ├── config/             # Конфигурирование
│   ├── container/          # Контейнер DI, хелперы
│   ├── db/                 # Абстракция БД
│   │   └── postgres/       # Реализация для PostgreSQL
│   ├── domain/             # Абстракция domain model, repository
│   ├── errs/               # Ошибки
│   ├── helper/             # helpers
│   ├── infra/              # Инфраструктура
│   │   ├── cache/          # Кэш
│   │   ├── metrics/        # Метрики
│   │   ├── pubsub/         # pub/sub шаблон
│   │   └── telemetry/      # Телеметрия (open tracing)
│   ├── logger/             # Логирование
│   ├── migration/          # Абстракция миграции данных
│   │   └── goose/          # Реализация миграции данных, библиотека goose
│   ├── repository/         # Базовые реализации репозитериев (CRUD/Owned), хелперы
│   │   ├── metrics/        # Базовые реализации обёрток метрик
│   │   └── trace/          # Базовые реализации обёрток телеметрии
│   ├── transport/          # транспорт
│   │   ├── brokers/        # абстракции работы с брокерами 
│   │   │   │ amqp/         # абстракции работы с брокером по протоколу AMQP
│   │   │   │   └ azure/    # реалищация работы с брокером, протокол AMQP (Azure/go-amqp)
│   │   │   └ kafka/        # реализация работы с брокером, kafka (segmentio/kafka-go)
│   │   ├── grpc/           # gRPC
│   │   │   └ interceptors/ # перехватчики
│   │   ├── http/           # HTTP
│   │   │   └ middleware/   # middleware
│   │   └── worker/         # воркеры
│   └── utils/              # утилиты
├── scripts/                # Вспомогательные скрипты
├── .gitignore              # Список файлов к пропуску для git 
├── .golangci.yml           # Конфигурация линтера golangci-lint 
├── .mockery.yml            # Конфигурация генератора mock файлов mockery 
├── Makefile                # Команды автоматизации (launch: make help)
├── LICENSE                 # Apache 2.0 лицензия 
├── go.mod                  # Зависимости
├── go.sum                  # Контрольные суммы
├── revive.toml             # Конфигурая линтера revive 
└── readme.md               # Документация проекта
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
* http with chi router and pipeline setup(tracing, metrics, logging, compress/decompress)
* gRPC with standard google library and pipeline setup (tracing, metrics)
* application configuration with viper library
* and application skeleton itself :-)

## Цепочка вызовов
### `transport -> facade -> use case -> repository -> lib/helper` и обратно
