package logger

import (
	"github.com/pressly/goose/v3"
)

// GooseLogger is implementation of goose.Logger interface
type GooseLogger struct {
	log Logger
}

var _ goose.Logger = (*GooseLogger)(nil)

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
