package main

import (
	"errors"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/app"
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
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
	var err error
	// 0. startup logger
	var startupLogger logger.Logger
	startupLogger, err = logger.NewZapLogger("info", "")
	if err != nil {
		panic(err)
	}
	defer startupLogger.Close()
	startupLogger = startupLogger.GetLogger("main")

	// 1. Загрузка конфигурации
	startupLogger.Info("init config")
	cfg, err := config.Load()
	if err != nil {
		startupLogger.Errorf("failed to load config: %v", err)

		panic(err)
	}

	// 2. Инициализация логгера
	startupLogger.Info("init logger")
	zapLogger, err := logger.NewZapLogger(cfg.Log.Level, cfg.Log.FilePath)
	if err != nil {
		startupLogger.Errorf("failed to init logger: %v", err)

		panic(err)
	}
	defer zapLogger.Close()

	// 3. Создание приложения
	startupLogger.Info("create application")
	application := app.NewApp(cfg, zapLogger)

	// 4. Инициализация приложения
	startupLogger.Info("init application")
	if err := application.Init(); err != nil {
		startupLogger.Errorf("failed application initialization [%v]", err)
		application.Close()

		panic(errs.NewCommonError("failed application initialization", err))
	}

	// 5. Запуск приложения
	startupLogger.Info("run application")
	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		application.Stop()
		startupLogger.Errorf("run application error [%v]", err)
	}

	// 6. Ожидание завершения приложения
	application.WaitForStop()

	// 7. Освобождение ресурсов
	startupLogger.Info("close application")
	application.Close()

	startupLogger.Info("shutdown application")
}
