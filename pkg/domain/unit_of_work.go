package domain

import (
	"context"
)

// UnitOfWork defines the pure domain interface for orchestrating atomic business transaction boundaries.
// It acts as a behavioral abstraction pattern (Unit of Work), shielding the high-level application Use Cases
// from explicit relational database mechanics, transaction states, or technical storage isolation parameters.
type UnitOfWork interface {
	// Execute encapsulates the execution of the provided domain business closure fn within an isolated atomic boundary.
	// The implementation must guarantee full rollback capability upon receiving a non-nil error state from the closure.
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}
