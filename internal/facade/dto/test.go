package dto

import (
	"time"
)

// Test представляет domain.Test model
// @Description Test DTO
type Test struct {
	ID           string    `json:"id,omitempty"`
	Code         string    `json:"code,omitempty"`
	Name         string    `json:"name,omitempty"`
	Description  string    `json:"description,omitempty"`
	RegisteredAt time.Time `json:"registered_at,omitzero"`
	UpdatedAt    time.Time `json:"updated_at,omitzero"`
} // @name Test
