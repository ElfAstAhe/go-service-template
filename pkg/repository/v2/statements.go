package v2

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Statement представляет собой строго типизированный инфраструктурный контейнер версии v2,
// инкапсулирующий SQL-запрос и соответствующую функцию-трансформер (Mapper) рекордов СУБД в Entity.
type Statement[E domain.Entity[ID], ID comparable] struct {
	SQL    string                  // Сырой текст SQL-инструкции или шаблона запроса
	Mapper EntityMapperFunc[E, ID] // Функция-маппер для безопасного преобразования результата выполнения в доменную модель
}

// NewStatement — фабричный конструктор для создания одиночной СУБД-инструкции со встроенным маппером.
func NewStatement[E domain.Entity[ID], ID comparable](
	sql string,
	mapper EntityMapperFunc[E, ID],
) *Statement[E, ID] {
	return &Statement[E, ID]{
		SQL:    sql,
		Mapper: mapper,
	}
}

// Statements инкапсулирует в себе внутренний реестр (мапу) именованных инструкций Statement[E, ID],
// обеспечивая централизованное хранение и валидацию SQL-команд для конкретного репозитория.
type Statements[E domain.Entity[ID], ID comparable] struct {
	statements map[string]*Statement[E, ID] // Внутреннее потокобезопасное (при условии read-only рантайма) хранилище команд СУБД
}

// NewStatements — фабричный конструктор реестра СУБД-инструкций текущего репозитория.
func NewStatements[E domain.Entity[ID], ID comparable]() *Statements[E, ID] {
	return &Statements[E, ID]{
		statements: make(map[string]*Statement[E, ID]),
	}
}

// Validate проверяет инварианты коллекции: реестр обязан содержать как минимум одну зарегистрированную SQL-команду.
func (s *Statements[E, ID]) Validate() error {
	if len(s.statements) == 0 {
		return errs.NewDalError("Statements[E, ID].Validate", "statements required", nil)
	}

	return nil
}

// Add осуществляет регистрацию или перезапись именованной СУБД-инструкции во внутреннем пуле.
func (s *Statements[E, ID]) Add(name string, action *Statement[E, ID]) {
	s.statements[name] = action
}

// Remove осуществляет удаление ранее зарегестрированного стейтмента
func (s *Statements[E, ID]) Remove(name string) {
	delete(s.statements, name)
}

// Has проверяет наличие зарегистрированной инструкции по её системному имени-маркеру.
func (s *Statements[E, ID]) Has(name string) bool {
	_, ok := s.statements[name]

	return ok
}

// Get извлекает указатель на объект инструкции Statement по имени. Возвращает DalError, если команда не найдена.
func (s *Statements[E, ID]) Get(name string) (*Statement[E, ID], error) {
	stmt, ok := s.statements[name]
	if !ok {
		return nil, errs.NewDalError("Statements[E, ID].GetMapper", fmt.Sprintf("statement mapper %s not found", name), nil)
	}

	return stmt, nil
}

// GetMapper — специализированный экспресс-геттер для прямого извлечения функции-трансформера по имени операции.
func (s *Statements[E, ID]) GetMapper(name string) (EntityMapperFunc[E, ID], error) {
	stmt, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	return stmt.Mapper, nil
}

// GetSQL — специализированный экспресс-геттер для извлечения сырой строки SQL-запроса по имени операции.
func (s *Statements[E, ID]) GetSQL(name string) (string, error) {
	stmt, err := s.Get(name)
	if err != nil {
		return "", err
	}

	return stmt.SQL, nil
}

// WithStatement добавляет элемент в виде fluent api
func (s *Statements[E, ID]) WithStatement(name string, stmt *Statement[E, ID]) *Statements[E, ID] {
	s.statements[name] = stmt

	return s
}
