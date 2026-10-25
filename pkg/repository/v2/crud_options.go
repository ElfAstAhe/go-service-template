package v2

import (
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// CRUDOption определяет сигнатуру функциональной опции (конфигуратора) для настройки параметров CRUD-репозитория.
type CRUDOption[E domain.Entity[ID], ID comparable] func(*CRUDOptions[E, ID])

// CRUDOptions инкапсулирует в себе полный набор инфраструктурных зависимостей, необходимых для рантайм-работы репозитория.
type CRUDOptions[E domain.Entity[ID], ID comparable] struct {
	Executor      db.Executor
	ErrorDecipher db.ErrorDecipher
	Statements    *Statements[E, ID]
	Hooks         []Hook[E, ID]
	TableName     string
	EntityName    string
}

// NewCRUDOptions — фабричный конструктор для создания пустого объекта конфигурации CRUDOptions.
func NewCRUDOptions[E domain.Entity[ID], ID comparable]() *CRUDOptions[E, ID] {
	return &CRUDOptions[E, ID]{}
}

// Validate выполняет декларативную проверку инвариантов конфигурации на предмет наличия обязательных зависимостей.
func (co *CRUDOptions[E, ID]) Validate() error {
	if utils.IsNil(co.Executor) {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "db executor required", nil)
	}
	if utils.IsNil(co.ErrorDecipher) {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "db error decipher required", nil)
	}
	if err := co.Statements.Validate(); err != nil {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "actions validation failed", err)
	}
	if utils.IsNil(co.Statements) {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "statements required", nil)
	}
	if strings.TrimSpace(co.TableName) == "" {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "table name required", nil)
	}
	if strings.TrimSpace(co.EntityName) == "" {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "entity name required", nil)
	}

	return nil
}

// ================================================================
// Fluent API
// ================================================================

// WithExecutor инжектирует в конфигурацию репозитория конкретный экземпляр исполнителя SQL-запросов (db.Executor).
func WithExecutor[E domain.Entity[ID], ID comparable](executor db.Executor) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		co.Executor = executor
	}
}

// WithErrorDecipher инжектирует компонент дешифрации нативных ошибок СУБД в типизированные ошибки DAL-слоя.
func WithErrorDecipher[E domain.Entity[ID], ID comparable](errorDecipher db.ErrorDecipher) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		co.ErrorDecipher = errorDecipher
	}
}

// WithStatement выполняет ленивую инициализацию реестра и добавляет в него именованную SQL-инструкцию со сканером.
func WithStatement[E domain.Entity[ID], ID comparable](name string, statement *Statement[E, ID]) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		if utils.IsNil(co.Statements) {
			co.Statements = NewStatements[E, ID]()
		}

		co.Statements.Add(name, statement)
	}
}

// WithHooks инициализирует реестр и регистрирует в конфигурации цепочку хуков и интерцепторов жизненного цикла сущности.
func WithHooks[E domain.Entity[ID], ID comparable](hooks ...Hook[E, ID]) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		co.Hooks = hooks
	}
}

// WithEntityInfo инжектирует информацию о таблице и сущности
func WithEntityInfo[E domain.Entity[ID], ID comparable](tableName, entityName string) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		co.TableName = tableName
		co.EntityName = entityName
	}
}
