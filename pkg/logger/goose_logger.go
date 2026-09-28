package logger

import (
	"github.com/pressly/goose/v3"
)

// GooseLogger выступает в роли инфраструктурного моста (Adapter), который перехватывает
// внутренние текстовые логи движка миграций Goose и транслирует их в наш структурированный ZapLogger.
type GooseLogger struct {
	log Logger // Ссылка на глобальный абстрактный интерфейс логгера фреймворка
}

// Гарантируем полное соответствие контракту goose.Logger на этапе компиляции
var _ goose.Logger = (*GooseLogger)(nil)

// NewGooseLogger — фабричный конструктор адаптера логирования миграций.
func NewGooseLogger(log Logger) *GooseLogger {
	return &GooseLogger{
		log: log,
	}
}

// Fatalf перехватывает фатальные ошибки Goose (например, синтаксический сбой в SQL-файле)
// и логирует их через системный уровень Errorf фреймворка.
func (g *GooseLogger) Fatalf(format string, v ...any) {
	g.log.Errorf(format, v...)
}

// Printf перехватывает информационные сообщения Goose (названия накатываемых файлов, статус транзакций)
// и аккуратно записывает их через уровень Infof.
func (g *GooseLogger) Printf(format string, v ...any) {
	g.log.Infof(format, v...)
}
