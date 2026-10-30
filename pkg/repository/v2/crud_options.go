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
	Executor         db.Executor                         // Database executor
	ErrorDecipher    db.ErrorDecipher                    // Database error decipher
	Hooks            []Hook[E, ID]                       // Hooks
	TableName        string                              // Table name info
	EntityName       string                              // Entity name info
	DriverName       string                              // database driver name (pgx, etc.)
	driverStatements map[string]StatementsBuilder[E, ID] // statements builder by db driver name
}

type StatementsBuilder[E domain.Entity[ID], ID comparable] func() *Statements[E, ID]

// NewCRUDOptions — фабричный конструктор для создания пустого объекта конфигурации CRUDOptions.
func NewCRUDOptions[E domain.Entity[ID], ID comparable]() *CRUDOptions[E, ID] {
	return &CRUDOptions[E, ID]{
		driverStatements: make(map[string]StatementsBuilder[E, ID]),
	}
}

// Validate выполняет декларативную проверку инвариантов конфигурации на предмет наличия обязательных зависимостей.
func (co *CRUDOptions[E, ID]) Validate() error {
	if utils.IsNil(co.Executor) {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "db executor required", nil)
	}
	if utils.IsNil(co.ErrorDecipher) {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "db error decipher required", nil)
	}
	if len(co.driverStatements) == 0 {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "driver statements builders required", nil)
	}
	if strings.TrimSpace(co.TableName) == "" {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "table name required", nil)
	}
	if strings.TrimSpace(co.EntityName) == "" {
		return errs.NewDalError("CRUDOptions[E, ID].Validate", "entity name required", nil)
	}

	return nil
}

func (co *CRUDOptions[E, ID]) GetStatements(driverName string) (*Statements[E, ID], error) {
	builder, ok := co.driverStatements[driverName]
	if !ok {
		return nil, errs.NewDalError("CRUDOptions[E, ID].GetStatements", "driver statements not found", nil)
	}

	return builder(), nil
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

// WithDriverName инжектирует database driver name для statements builder
func WithDriverName[E domain.Entity[ID], ID comparable](name string) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		co.DriverName = name
	}
}

// WithStatementsBuilder ленивая инициализация карты билдеров Statements
func WithStatementsBuilder[E domain.Entity[ID], ID comparable](driverName string, builder StatementsBuilder[E, ID]) CRUDOption[E, ID] {
	return func(co *CRUDOptions[E, ID]) {
		if utils.IsNil(co.driverStatements) {
			co.driverStatements = make(map[string]StatementsBuilder[E, ID])
		}
		co.driverStatements[driverName] = builder
	}
}
