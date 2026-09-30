package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// UnitOfWork bridges the core application business logic layer with the platform's database transactions.
// It implements the domain.UnitOfWork interface, enforcing clean architectural boundaries via dependency inversion.
type UnitOfWork struct {
	tm         db.TransactionManager  // Infrastructure transaction coordinator from the platform core
	tmo        *db.TransactionOptions // Optional custom explicit transaction configurations
	defaultTMO *db.TransactionOptions // Fallback default safe transaction configuration passport
}

// Compile-time interface compliance verification
var _ domain.UnitOfWork = (*UnitOfWork)(nil)

// NewUnitOfWork creates a new UnitOfWork instance, initializing default transactional fallback configurations.
func NewUnitOfWork(tm db.TransactionManager, tmo *db.TransactionOptions) *UnitOfWork {
	return &UnitOfWork{
		tm:         tm,
		tmo:        tmo,
		defaultTMO: buildDefaultTransactionOptions(),
	}
}

// Execute encapsulates the operational business logic closure inside a platform-managed database transaction context.
func (uow *UnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	return uow.tm.WithinTransaction(ctx, uow.getTransactionOptionsOrDefault(), fn)
}

// getTransactionOptionsOrDefault evaluates if explicit custom options are active; otherwise, yields fallback defaults.
func (uow *UnitOfWork) getTransactionOptionsOrDefault() *db.TransactionOptions {
	if !utils.IsNil(uow.tmo) {
		return uow.tmo
	}

	return uow.defaultTMO
}

// buildDefaultTransactionOptions structures standard database default isolation constraints (ANSI SQL).
func buildDefaultTransactionOptions() *db.TransactionOptions {
	return &db.TransactionOptions{
		Isolation: db.LevelDefault,
		ReadOnly:  false,
	}
}
