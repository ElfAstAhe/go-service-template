package v2

import "github.com/ElfAstAhe/go-service-template/pkg/domain"

// dummyModel — фейковая доменная модель для изоляции дженериков в тестах.
type dummyModel struct{ ID string }

func (d *dummyModel) GetID() string         { return d.ID }
func (d *dummyModel) SetID(id string)       { d.ID = id }
func (d *dummyModel) IsExists() bool        { return d.ID != "" }
func (d *dummyModel) BeforeCreate() error   { return nil }
func (d *dummyModel) BeforeChange() error   { return nil }
func (d *dummyModel) ValidateCreate() error { return nil }
func (d *dummyModel) ValidateChange() error { return nil }

var _ domain.Entity[string] = (*dummyModel)(nil)
