package domain

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

type Test2Repository interface {
	domain.CRUDRepository[*Test2, string]

	FindByCode(ctx context.Context, code string) (*Test2, error)
}
