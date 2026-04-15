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

.PHONY: build run test clean

# Генерация gRPC кода
gen-proto:
	mkdir -p $(PROTO_OUT)
	protoc --proto_path=$(PROTO_PATH) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--go_opt=default_api_level=API_OPAQUE \
		$(PROTO_PATH)/*.proto

# Генерация swagger
gen-swagger:
	swag init \
		-g $(SERVER_BUILD_DIR)/main.go \
		--parseDependency \
		--parseInternal \
		--exclude ./pkg/api \
		-o docs \
		--parseDepth 3

# Генерация http client
gen-http-client:
#	oapi-codegen -package client -generate client docs/swagger.json > pkg/client/rest/api_client.gen.go
	mkdir -p $(OPEN_API_OUT)
	swagger generate client -f ./docs/swagger.json -A go-service-template -t $(OPEN_API_OUT)

# Генерирует моки для интерфейсов в указанной папке, см. {project_root}/.mockery.yml конфиг
gen-mocks:
	mockery

# Сборка всего с прокидыванием переменных
build: gen-proto gen-swagger gen-http-client gen-mocks
	go build -ldflags \
	"-X '$(MODULE_NAME)/internal/config.AppVersion=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/config.AppBuildTime=$(BUILD_TIME)'" \
	-o ./bin/$(SERVER_BINARY_NAME) $(SERVER_BUILD_DIR)/main.go

# Сборка проекта с прокидыванием переменных
build-only: gen-proto gen-swagger gen-http-client
	go build -ldflags \
	"-X '$(MODULE_NAME)/internal/config.AppVersion=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/config.AppBuildTime=$(BUILD_TIME)'" \
	-o ./bin/$(SERVER_BINARY_NAME) $(SERVER_BUILD_DIR)/main.go

# Запуск проекта (сначала соберет, потом запустит)
run: build
	./bin/$(SERVER_BINARY_NAME) \
		--db-driver "postgres" \
		--db-dsn "postgres://test:password@localhost:5432/test?sslmode=disable&search_path=example_service"

# Запуск тестов
test:
	go test -v ./...

# Запуск бенчмарков (сюда добавляем все вызовы) или разные параметры под один пакет
bench:
	go test -bench=BenchmarkManager_FullCycle -benchmem ./pkg/infra/cache/test/...

# Запуск static check
static-check:
	staticcheck $$(go list ./... | grep -vE "pkg/api")

# Очистка бинарников
clean:
	rm -rf ./bin/*
