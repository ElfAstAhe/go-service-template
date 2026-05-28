package migration

import (
	"context"
)

// Migrator is simple database migrator interface
type Migrator interface {
	Initialize() error
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}
