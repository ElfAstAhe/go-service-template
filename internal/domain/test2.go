package domain

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Test2 struct {
	ID          string
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
	ModifiedAt  time.Time
}

var _ domain.Entity[string] = (*Test2)(nil)

func NewEmptyTest2() *Test2 {
	return &Test2{}
}

func NewTest2(id, code, name, description string) *Test2 {
	return &Test2{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: description,
	}
}

func (t2 *Test2) GetID() string {
	return t2.ID
}

func (t2 *Test2) SetID(id string) {
	t2.ID = id
}

func (t2 *Test2) IsExists() bool {
	return strings.TrimSpace(t2.ID) != ""
}

func (t2 *Test2) BeforeCreate() error {
	if err := domain.AssignUUIDv7(t2); err != nil {
		return errs.NewBllError("Test2.BeforeCreate", "assign uuid v7 failed", err)
	}

	if t2.CreatedAt.IsZero() {
		t2.CreatedAt = time.Now()
	}
	t2.ModifiedAt = time.Now()

	return nil
}

func (t2 *Test2) BeforeChange() error {
	t2.ModifiedAt = time.Now()

	return nil
}

func (t2 *Test2) ValidateCreate() error {
	if strings.TrimSpace(t2.ID) != "" {
		return errs.NewBllValidateError("Test2.ValidateCreate", "ID should be empty", nil)
	}
	if strings.TrimSpace(t2.Code) == "" {
		return errs.NewBllValidateError("Test2.ValidateCreate", "Code should be set", nil)
	}

	return nil
}

func (t2 *Test2) ValidateChange() error {
	if strings.TrimSpace(t2.ID) == "" {
		return errs.NewBllValidateError("Test2.ValidateChange", "ID should be set", nil)
	}
	if strings.TrimSpace(t2.Code) == "" {
		return errs.NewBllValidateError("Test2.ValidateChange", "Code should be set", nil)
	}

	return nil
}
