package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger реализует абстрактный интерфейс Logger фреймворка, оборачивая высокопроизводительный
// структурированный логгер компании Uber (go.uber.org/zap).
type ZapLogger struct {
	logger *zap.Logger // Экземпляр нативного структурированного ядра Zap
}

// Гарантируем соответствие контракту Logger на этапе компиляции
var _ Logger = (*ZapLogger)(nil)

// NewStartupZapLogger создает легковесный логгер для этапа стартапа приложения.
// Выводит логи только в консоль на уровне INFO, пока основные файлы конфигурации еще не прочитаны Viper-ом.
func NewStartupZapLogger() *ZapLogger {
	zapLevel := zap.NewAtomicLevelAt(zap.InfoLevel)

	return &ZapLogger{
		// AddCallerSkip(1) смещает указатель стека, чтобы логгер показывал место вызова в доменном слое
		logger: zap.New(newConsoleZapCore(zapLevel), zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}

// NewZapLogger — основной конструктор промышленного логгера.
// Реализует паттерн Tee (Мультиплексор): параллельно транслирует логи в красивом человекочитаемом виде
// в Console (stdout) и в структурированном JSON-формате в файл на диске с поддержкой автоматической ротации.
func NewZapLogger(level string, filePath string) (*ZapLogger, error) {
	zapLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	consoleCore := newConsoleZapCore(zapLevel)
	core := zapcore.NewTee(consoleCore)

	// Если путь к файлу передан — подключаем дисковое JSON-логирование с ротацией
	if strings.TrimSpace(filePath) != "" {
		fileCore := newFileZapCore(zapLevel, filePath)
		core = zapcore.NewTee(consoleCore, fileCore)
	}

	res := &ZapLogger{}
	res.logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return res, nil
}

// newConsoleZapCore собирает ядро для вывода логов в консоль разработчика (Development-режим).
func newConsoleZapCore(level zap.AtomicLevel) zapcore.Core {
	stdOut := zapcore.AddSync(os.Stdout)

	encConf := zap.NewDevelopmentEncoderConfig()
	// Включает цветную кодировку уровней логов (INFO - зеленый, ERROR - красный) для удобства чтения
	encConf.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encConf)
	res := zapcore.NewCore(consoleEncoder, stdOut, level)

	return res
}

// newFileZapCore собирает ядро для вывода логов в ротируемый JSON-файл (Production-режим).
func newFileZapCore(level zap.AtomicLevel, filePath string) zapcore.Core {
	// lumberjack.Logger защищает диски от переполнения
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    10, // Ротация файла при достижении 10 Мегабайт
		MaxBackups: 3,  // Хранить не более 3 архивных файлов
		MaxAge:     7,  // Удалять архивы старше 7 дней
	})

	encConf := zap.NewProductionEncoderConfig()
	encConf.TimeKey = "timestamp"
	// Приведение меток времени к каноничному стандарту ISO8601 для корректной индексации базами (Loki)
	encConf.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(encConf)
	res := zapcore.NewCore(fileEncoder, file, level)

	return res
}

// Close принудительно сбрасывает все накопленные в буферах Zap-пакеты ввода-вывода (Flush).
func (zl *ZapLogger) Close() error {
	// Игнорируем специфичные системные ошибки Sync() при выводе в stdout/stderr на Linux
	_ = zl.logger.Sync()
	return nil
}

// ====================================================================
// Реализация методов интерфейса Logger через SugaredLogger (Сахарный слой)
// ====================================================================

// Error записывает сообщение уровня ошибки в структурированный лог.
func (zl *ZapLogger) Error(args ...any) { zl.logger.Sugar().Error(args...) }

// Errorf записывает форматированное сообщение уровня ошибки согласно переданному шаблону.
func (zl *ZapLogger) Errorf(format string, args ...any) { zl.logger.Sugar().Errorf(format, args...) }

// ErrorW записывает сообщение уровня ошибки, обогащенное парами произвольных ключ-значений (структурированный контекст).
func (zl *ZapLogger) ErrorW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Errorw(msg, keysAndValues...)
}

// Warn записывает сообщение уровня предупреждения в структурированный лог.
func (zl *ZapLogger) Warn(args ...any) { zl.logger.Sugar().Warn(args...) }

// Warnf записывает форматированное сообщение уровня предупреждения согласно переданному шаблону.
func (zl *ZapLogger) Warnf(format string, args ...any) { zl.logger.Sugar().Warnf(format, args...) }

// WarnW записывает сообщение уровня предупреждения, обогащенное парами произвольных ключ-значений.
func (zl *ZapLogger) WarnW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Warnw(msg, keysAndValues...)
}

// Info записывает информационное сообщение общего назначения в структурированный лог.
func (zl *ZapLogger) Info(args ...any) { zl.logger.Sugar().Info(args...) }

// Infof записывает форматированное информационное сообщение согласно переданному шаблону.
func (zl *ZapLogger) Infof(format string, args ...any) { zl.logger.Sugar().Infof(format, args...) }

// InfoW записывает информационное сообщение, обогащенное парами произвольных ключ-значений.
func (zl *ZapLogger) InfoW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Infow(msg, keysAndValues...)
}

// Debug записывает сообщение уровня отладки (Development/Tracing) в структурированный лог.
func (zl *ZapLogger) Debug(args ...any) { zl.logger.Sugar().Debug(args...) }

// Debugf записывает форматированное сообщение уровня отладки согласно переданному шаблону.
func (zl *ZapLogger) Debugf(format string, args ...any) { zl.logger.Sugar().Debugf(format, args...) }

// DebugW записывает сообщение уровня отладки, обогащенное парами произвольных ключ-значений.
func (zl *ZapLogger) DebugW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Debugw(msg, keysAndValues...)
}

// GetLogger порождает дочернее изолированное контекстное плечо логгера (Sub-Logger).
// Прикрепляет ключ "childEntry" ко всем последующим цепочкам логов текущего компонента.
func (zl *ZapLogger) GetLogger(logicEntry string) Logger {
	return &ZapLogger{
		logger: zl.logger.With(zap.String("childEntry", logicEntry)),
	}
}
