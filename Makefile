# Переменные для сборки
PROTO_PATH=api/proto/example-service/v1
PROTO_OUT=pkg/api/grpc/example/v1
OPEN_API_OUT=pkg/api/http/example/v1
MODULE_NAME=github.com/ElfAstAhe/go-service-template
SERVER_BINARY_NAME=example-service
SERVER_BUILD_DIR=./cmd/example-service
VERSION=1.0.0
BUILD_TIME=$(shell date +'%Y/%m/%d_%H:%M:%S')
STAGE=DEV

.PHONY: gen-proto gen-swagger gen-http-client gen-mocks gen-sources build build-only run test bench static-check lint clean update-deps
# Показывает это руководство (выполняется по умолчанию)
help:
	@echo "Доступные команды для сборки и тестирования:"
	@echo "------------------------------------------------------------------------"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo "------------------------------------------------------------------------"

# Генерация gRPC кода
gen-proto: ## Сгенерировать gRPC код (Go & gRPC) из Protobuf файлов
	mkdir -p $(PROTO_OUT)
	protoc --proto_path=$(PROTO_PATH) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--go_opt=default_api_level=API_OPAQUE \
		$(PROTO_PATH)/*.proto

# Генерация swagger
gen-swagger: ## Сгенерировать Swagger-документацию (swag init)
	swag init \
		-g $(SERVER_BUILD_DIR)/main.go \
		--parseDependency \
		--parseInternal \
		--exclude ./pkg/api \
		-o docs \
		--parseDepth 3

# Генерация http client
gen-http-client: ## Сгенерировать HTTP-клиент на основе swagger.json
#	oapi-codegen -package client -generate client docs/swagger.json > pkg/client/rest/api_client.gen.go
	mkdir -p $(OPEN_API_OUT)
	swagger generate client -f ./docs/swagger.json -A go-service-template -t $(OPEN_API_OUT)

# Генерирует моки для интерфейсов в указанной папке, см. {project_root}/.mockery.yml конфиг
gen-mocks: ## Сгенерировать моки для интерфейсов (mockery)
	mockery

# Генерирует исходники
gen-sources: ## Сгенерировать динамические исходники проекта (например, таймзоны)
	go run ./cmd/gen-tz/main.go

# Сборка всего с прокидыванием переменных
build: gen-sources gen-proto gen-swagger gen-http-client gen-mocks ## Полная сборка: генерация всего кода + компиляция бинарника
	go build -ldflags \
	"-X '$(MODULE_NAME)/internal/config.AppVersion=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/config.AppBuildTime=$(BUILD_TIME)'" \
	-o ./bin/$(SERVER_BINARY_NAME) $(SERVER_BUILD_DIR)/main.go

# Сборка проекта с прокидыванием переменных
build-only: gen-sources gen-proto gen-swagger gen-http-client ## Быстрая сборка: компиляция бинарника (без перегенерации моков)
	go build -ldflags \
	"-X '$(MODULE_NAME)/internal/config.AppVersion=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/config.AppBuildTime=$(BUILD_TIME)'" \
	-o ./bin/$(SERVER_BINARY_NAME) $(SERVER_BUILD_DIR)/main.go

# Запуск проекта (сначала соберет, потом запустит)
run: build ## Собрать проект и запустить бинарник с локальными флагами (БД, логи)
	./bin/$(SERVER_BINARY_NAME) \
		--db-driver "postgres" \
		--db-dsn "postgres://test:password@localhost:5432/test?sslmode=disable&search_path=example_service" \
		--log-level "DEBUG"

# Запуск тестов
test: gen-sources gen-proto gen-mocks ## Запустить модульные и интеграционные тесты проекта
	go test -v ./...

# Запуск бенчмарков (сюда добавляем все вызовы) или разные параметры под один пакет
bench: gen-sources gen-proto gen-mocks ## Запустить кэш-бенчмарки и утилиты с замером памяти
	go test -bench=BenchmarkManager_FullCycle -benchmem ./pkg/infra/cache/test/...
	go test -bench=. -benchmem ./pkg/utils/...

# Запуск static check
static-check: ## Запустить статический анализ кода (пропуская автогенерируемый pkg/api)
	staticcheck $$(go list ./... | grep -vE "pkg/api|cmd/gen-tz")

# Запуск линтера
lint: ## Запустить линтер revive (пропуская автогенерируемый код)
	revive $$(go list ./... | grep -vE "pkg/api|cmd/gen-tz")

# Очистка бинарников
clean: ## Очистить скомпилированные файлы из папки ./bin
	rm -rf ./bin/*

# обновление зависимостей
update-deps: ## Принудительно обновить и скачать все Go-зависимости проекта
	go get -u -x all

#