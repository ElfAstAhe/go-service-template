package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/app"
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// @title           Example Service API
// @version         1.0
// @description     Тестовый сервис
// @termsOfService  Free use

// @contact.name   API Support
// @contact.url    https://github.com/ElfAstAhe/go-service-template
// @contact.email  elf.ast.ahe@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org

// @BasePath  /
//
//goland:noinspection GoUnhandledErrorResult
func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Инициализация логгера
	zapLogger, err := logger.NewZapLogger(cfg.Log.Level, cfg.Log.FilePath)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Close()

	// 3. Создание и инициализация приложения
	application := app.NewApp(cfg, zapLogger)

	// app initialization
	zapLogger.Info("app init")
	if err := application.Init(); err != nil {
		zapLogger.Errorf("app init failed [%v]", err)
		application.Close()

		panic(fmt.Errorf("app initialization failed: %v", err))
	}

	// app run
	zapLogger.Info("app run")
	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		application.Stop()
		zapLogger.Errorf("app run error [%v]", err)
	}

	application.WaitForStop()

	// app close
	zapLogger.Info("app close")
	if err := application.Close(); err != nil {
		zapLogger.Errorf("app close error [%v]", err)

		panic(fmt.Errorf("app close failed: %v", err))
	}

	zapLogger.Info("app shutdown")
}
