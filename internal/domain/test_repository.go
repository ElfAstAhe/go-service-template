package domain

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

type TestRepository interface {
	domain.CrudRepository[*Test, string]

	FindByCode(ctx context.Context, code string) (*Test, error)
}
