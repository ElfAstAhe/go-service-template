package facade

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
)

type Test2Facade interface {
	Get(ctx context.Context, id string) (*dto.Test2, error)
	GetByCode(ctx context.Context, code string) (*dto.Test2, error)
	List(ctx context.Context, limit int, offset int) ([]*dto.Test2, error)
	Create(ctx context.Context, dto *dto.Test2) (*dto.Test2, error)
	Change(ctx context.Context, id string, dto *dto.Test2) (*dto.Test2, error)
	Delete(ctx context.Context, id string) error
}

type Test2 struct {
}

var _ Test2Facade = (*Test2)(nil)

func NewTest2() *Test2 {
	return &Test2{}
}

func (t2f *Test2) Get(ctx context.Context, id string) (*dto.Test2, error) {
	//TODO implement me
	panic("implement me")
}

func (t2f *Test2) GetByCode(ctx context.Context, code string) (*dto.Test2, error) {
	//TODO implement me
	panic("implement me")
}

func (t2f *Test2) List(ctx context.Context, limit int, offset int) ([]*dto.Test2, error) {
	//TODO implement me
	panic("implement me")
}

func (t2f *Test2) Create(ctx context.Context, instance *dto.Test2) (*dto.Test2, error) {
	//TODO implement me
	panic("implement me")
}

func (t2f *Test2) Change(ctx context.Context, id string, instance *dto.Test2) (*dto.Test2, error) {
	//TODO implement me
	panic("implement me")
}

func (t2f *Test2) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
