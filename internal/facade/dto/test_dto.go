package dto

import (
	"time"
)

type TestDTO struct {
	ID           string    `json:"id,omitempty"`
	Code         string    `json:"code,omitempty"`
	Name         string    `json:"name,omitempty"`
	Description  string    `json:"description,omitempty"`
	RegisteredAt time.Time `json:"registered_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
