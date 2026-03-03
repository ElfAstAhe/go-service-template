package main

import (
	"errors"
	"log"
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
	// 1. Загрузка конфигурации
	log.Println("MAIN: init config")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("MAIN: failed to load config: %v", err)
	}

	// 2. Инициализация логгера
	log.Println("MAIN: init logger")
	zapLogger, err := logger.NewZapLogger(cfg.Log.Level, cfg.Log.FilePath)
	if err != nil {
		log.Fatalf("MAIN: failed to init logger: %v", err)
	}
	defer zapLogger.Close()

	// 3. Создание приложения
	zapLogger.Info("MAIN: create application")
	application := app.NewApp(cfg, zapLogger)

	// 4. Инициализация приложения
	zapLogger.Info("MAIN: init application")
	if err := application.Init(); err != nil {
		zapLogger.Errorf("MAIN: failed application initialization [%v]", err)
		application.Close()

		panic(errs.NewCommonError("MAIN: failed application initialization", err))
	}

	// 5. Запуск приложения
	zapLogger.Info("MAIN: run application")
	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		application.Stop()
		zapLogger.Errorf("MAIN: run application error [%v]", err)
	}

	// 6. Ожидание завершения приложения
	application.WaitForStop()

	// 7. Освобождение ресурсов
	zapLogger.Info("MAIN: close application")
	application.Close()

	zapLogger.Info("MAIN: shutdown application")
}
