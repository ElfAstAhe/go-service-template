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
		defer application.Close()

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
