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
	Executor           db.Executor                             // Database executor
	ErrorDecipher      db.ErrorDecipher                        // Database error decipher
	Statements         *Statements[E, ID]                      // Statements
	Hooks              []Hook[E, ID]                           // Hooks
	TableName          string                                  // table name info
	EntityName         string                                  // entity name info
	DriverName         string                                  // current usable database driver
	StatementsBuilders map[string]CRUDStatementsBuilder[E, ID] // statements builder by database driver name
}

type CRUDStatementsBuilder[E domain.Entity[ID], ID comparable] func() []CRUDOption[E, ID]

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
	if IsStatementsEmpty(co.Statements) && len(co.StatementsBuilders) == 0 {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "statements or statements builder required", nil)
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

// WithDriverName индектирует database driver name for statements builder
func WithDriverName[E domain.Entity[ID], ID comparable](name string) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		co.DriverName = name
	}
}

// WithStatementsBuilder ленивая инициализация карты билдеров Statements
func WithStatementsBuilder[E domain.Entity[ID], ID comparable](driverName string, builder CRUDStatementsBuilder[E, ID]) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		if utils.IsNil(co.StatementsBuilders) {
			co.StatementsBuilders = make(map[string]CRUDStatementsBuilder[E, ID])
		}
		co.StatementsBuilders[driverName] = builder
	}
}
