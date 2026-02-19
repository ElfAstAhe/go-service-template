package logger

type Logger interface {
	// Структурированные логи
	// Позволяют писать: log.InfoW("user created", "id", user.ID, "ip", ip)
	InfoW(msg string, keysAndValues ...any)
	WarnW(msg string, keysAndValues ...any)
	ErrorW(msg string, keysAndValues ...any)
	DebugW(msg string, keysAndValues ...any)

	// простые логи
	Error(args ...any)
	Warn(args ...any)
	Info(args ...any)
	Debug(args ...any)

	// формтированные логи
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Debugf(format string, args ...any)

	GetLogger(logicEntry string) Logger

	Close() error
}
