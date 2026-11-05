package v2

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// dummyModel — фейковая доменная модель для изоляции дженериков в тестах.
type dummyModel struct{ ID string }

var _ domain.Entity[string] = (*dummyModel)(nil)

// GetID возвращает ID
func (d *dummyModel) GetID() string { return d.ID }

// SetID устанавливает ID
func (d *dummyModel) SetID(id string) { d.ID = id }

// IsExists возвращает признак наличия записи в СУБД (проверка наличия ID)
func (d *dummyModel) IsExists() bool { return d.ID != "" }

// BeforeCreate метод жизненного цикла, вызывается перед созданием записи
func (d *dummyModel) BeforeCreate() error { return nil }

// BeforeChange метод жизненного цикла, вызывается перед изменением записи
func (d *dummyModel) BeforeChange() error { return nil }

// ValidateCreate метод проверки сущности (атрибутов, etc) перед созданием записи
func (d *dummyModel) ValidateCreate() error { return nil }

// ValidateChange метод проверки сущности (атрибутов, etc) перед изменением записи
func (d *dummyModel) ValidateChange() error { return nil }
