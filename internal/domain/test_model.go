package domain

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
)

type Test struct {
	ID          string
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
	ModifiedAt  time.Time
}

func NewEmptyTest() *Test {
	return &Test{}
}

func NewTest(id string, code string, name string, description string, createdAt time.Time, modifiedAt time.Time) *Test {
	return &Test{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		ModifiedAt:  modifiedAt,
	}
}

func (t *Test) GetID() string {
	return t.ID
}

func (t *Test) SetID(id string) {
	t.ID = id
}

func (t *Test) IsExists() bool {
	return t.ID != ""
}

func (t *Test) BeforeCreate() error {
	t.CreatedAt = time.Now()
	t.ModifiedAt = time.Now()

	return nil
}

func (t *Test) BeforeChange() error {
	t.ModifiedAt = time.Now()

	return nil
}

func (t *Test) ValidateCreate() error {
	if t.ID != "" {
		return errs.NewBllValidateError("", "ID should be empty", nil)
	}
	if t.Code == "" {
		return errs.NewBllValidateError("", "Code should be set", nil)
	}

	return nil
}

func (t *Test) ValidateChange() error {
	if t.ID == "" {
		return errs.NewBllValidateError("", "ID should be set", nil)
	}
	if t.Code == "" {
		return errs.NewBllValidateError("", "Code should be set", nil)
	}

	return nil
}
