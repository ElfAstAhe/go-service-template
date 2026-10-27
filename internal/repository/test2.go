package repository

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
    v2 "github.com/ElfAstAhe/go-service-template/pkg/repository/v2"
)

type Test2 struct {
	*v2.CRUD[*domain.Test2, string]
}

var _ pkgdom.CRUDRepository[*domain.Test2, string] = (*Test2)(nil)
var _ domain.Test2Repository = (*Test2)(nil)

func NewTest2(executor db.Executor, errDecipher db.ErrorDecipher) (*Test2, error) {
	// instance
	res := &Test2{}
	// embed
	embed, err := v2.NewCRUD[*domain.Test2, string](
		v2.WithEntityInfo[*domain.Test2, string]("test2", "Test2"),
		v2.WithErrorDecipher[*domain.Test2, string](errDecipher),
		v2.WithExecutor[*domain.Test2, string](executor),
		v2.WithHooks[*domain.Test2, string](
			v2.WithHookCreator[*domain.Test2, string](res.creator),
			v2.WithHookChanger[*domain.Test2, string](res.changer),
		),
		v2.WithStatement[*domain.Test2, string](v2.StatementFind, v2.NewStatement[*domain.Test2, string](sqlTest2))
	)
	if err != nil {
		return nil, errs.NewDalError("NewTest2", "create embed CRUD failed", err)
	}
	// setup
	res.CRUD = embed

	return res, nil
}

func (t2 *Test2) creator(ctx context.Context, querier db.Querier, entity *domain.Test2, params... any) (v2.Scannable, error) {

}

func (t2 *Test2) changer(ctx context.Context, querier db.Querier, entity *domain.Test2, params... any) (v2.Scannable, error) {

}

func (t2 *Test2) mapDefault(ctx context.Context, scannable v2.Scannable, params... any) (*domain.Test2, error) {

}