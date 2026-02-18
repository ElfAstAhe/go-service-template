package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewStartupZapLogger() *ZapLogger {
	zapLevel := zap.NewAtomicLevelAt(zap.InfoLevel)

	return &ZapLogger{
		logger: zap.New(newConsoleZapCore(zapLevel), zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}

func NewZapLogger(level string, filePath string) (*ZapLogger, error) {
	zapLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	consoleCore := newConsoleZapCore(zapLevel)
	core := zapcore.NewTee(consoleCore)
	if strings.TrimSpace(filePath) != "" {
		fileCore := newFileZapCore(zapLevel, filePath)

		core = zapcore.NewTee(consoleCore, fileCore)
	}

	res := &ZapLogger{}

	res.logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return res, nil
}

func newConsoleZapCore(level zap.AtomicLevel) zapcore.Core {
	stdOut := zapcore.AddSync(os.Stdout)

	encConf := zap.NewDevelopmentEncoderConfig()
	encConf.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encConf)

	res := zapcore.NewCore(consoleEncoder, stdOut, level)

	return res
}

func newFileZapCore(level zap.AtomicLevel, filePath string) zapcore.Core {
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
	})

	encConf := zap.NewProductionEncoderConfig()
	encConf.TimeKey = "timestamp"
	encConf.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(encConf)

	res := zapcore.NewCore(fileEncoder, file, level)

	return res
}

// Closer

func (zl *ZapLogger) Close() error {
	return zl.logger.Sync()
}

// Logger

func (zl *ZapLogger) Error(args ...any) {
	zl.logger.Sugar().Error(args...)
}

func (zl *ZapLogger) Errorf(format string, args ...any) {
	zl.logger.Sugar().Errorf(format, args...)
}

func (zl *ZapLogger) ErrorW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Errorw(msg, keysAndValues...)
}

func (zl *ZapLogger) Warn(args ...any) {
	zl.logger.Sugar().Warn(args...)
}

func (zl *ZapLogger) Warnf(format string, args ...any) {
	zl.logger.Sugar().Warnf(format, args...)
}

func (zl *ZapLogger) WarnW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Warnw(msg, keysAndValues...)
}

func (zl *ZapLogger) Info(args ...any) {
	zl.logger.Sugar().Info(args...)
}

func (zl *ZapLogger) Infof(format string, args ...any) {
	zl.logger.Sugar().Infof(format, args...)
}

func (zl *ZapLogger) InfoW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Infow(msg, keysAndValues...)
}

func (zl *ZapLogger) Debug(args ...any) {
	zl.logger.Sugar().Debug(args...)
}

func (zl *ZapLogger) Debugf(format string, args ...any) {
	zl.logger.Sugar().Debugf(format, args...)
}

func (zl *ZapLogger) DebugW(msg string, keysAndValues ...any) {
	zl.logger.Sugar().Debugw(msg, keysAndValues...)
}

func (zl *ZapLogger) GetLogger(logicEntry string) Logger {
	return &ZapLogger{
		logger: zl.logger.With(zap.String("childEntry", logicEntry)),
	}
}
