package dto

import (
	"time"
)

// Test2 represent domain.Test2 model
// @Description Test2 DTO
type Test2 struct {
	ID          string    `json:"id,omitempty" yaml:"id,omitempty"`
	Code        string    `json:"code,omitempty" yaml:"code,omitempty"`
	Name        string    `json:"name,omitempty" yaml:"name,omitempty"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitzero" yaml:"created_at,omitzero"`
	ModifiedAt  time.Time `json:"modified_at,omitzero" yaml:"modified_at,omitzero"`
} // @name Test2
