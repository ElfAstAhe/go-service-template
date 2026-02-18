package logger

// GooseLogger is implementation of goose.Logger interface
type GooseLogger struct {
	log Logger
}

func NewGooseLogger(log Logger) *GooseLogger {
	return &GooseLogger{
		log: log,
	}
}

// Logger

func (g *GooseLogger) Fatalf(format string, v ...interface{}) {
	g.log.Errorf(format, v...)
}

func (g *GooseLogger) Printf(format string, v ...interface{}) {
	g.log.Infof(format, v...)
}

// ==========
