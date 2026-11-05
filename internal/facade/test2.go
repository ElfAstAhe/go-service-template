package facade

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/ElfAstAhe/go-service-template/internal/facade/mapper"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
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
	getUC     usecase.Test2GetUseCase
	getByCode usecase.Test2GetByCodeUseCase
	listUC    usecase.Test2ListUseCase
	saveUC    usecase.Test2SaveUseCase
	deleteUC  usecase.Test2DeleteUseCase
}

var _ Test2Facade = (*Test2)(nil)

// NewTest2 фабричный метод создания экземпляра Test2
func NewTest2(
	getUC usecase.Test2GetUseCase,
	getByCodeUC usecase.Test2GetByCodeUseCase,
	listUC usecase.Test2ListUseCase,
	saveUC usecase.Test2SaveUseCase,
	deleteUC usecase.Test2DeleteUseCase,
) *Test2 {
	return &Test2{
		getUC:     getUC,
		getByCode: getByCodeUC,
		listUC:    listUC,
		saveUC:    saveUC,
		deleteUC:  deleteUC,
	}
}

func (t2f *Test2) Get(ctx context.Context, id string) (*dto.Test2, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errs.NewInvalidArgumentError("id", "required")
	}

	model, err := t2f.getUC.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.MapTest2ModelToDTO(model), nil
}

func (t2f *Test2) GetByCode(ctx context.Context, code string) (*dto.Test2, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errs.NewInvalidArgumentError("code", "required")
	}

	model, err := t2f.getByCode.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	return mapper.MapTest2ModelToDTO(model), nil
}

func (t2f *Test2) List(ctx context.Context, limit int, offset int) ([]*dto.Test2, error) {
	if err := t2f.validateList(limit, offset); err != nil {
		return nil, err
	}
	models, err := t2f.listUC.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return mapper.MapTest2ModelsToDTO(models), nil
}

func (t2f *Test2) validateList(limit, offset int) error {
	if limit <= 0 {
		return errs.NewInvalidArgumentError("limit", "must be greater than 0")
	}
	if offset < 0 {
		return errs.NewInvalidArgumentError("offset", "must be greater or equal than 0")
	}
	if limit > DefaultMaxListLimit {
		return errs.NewInvalidArgumentError("limit", "must be less or equal than 1000")
	}

	return nil
}

func (t2f *Test2) Create(ctx context.Context, instance *dto.Test2) (*dto.Test2, error) {
	if utils.IsNil(instance) {
		return nil, errs.NewInvalidArgumentError("instance", "required")
	}

	model := mapper.MapTest2DTOToModel(instance)
	model.ID = ""

	var err error
	model, err = t2f.saveUC.Save(ctx, model)
	if err != nil {
		return nil, err
	}

	return mapper.MapTest2ModelToDTO(model), nil
}

func (t2f *Test2) Change(ctx context.Context, id string, instance *dto.Test2) (*dto.Test2, error) {
	if utils.IsNil(instance) {
		return nil, errs.NewInvalidArgumentError("instance", "required")
	}

	model := mapper.MapTest2DTOToModel(instance)
	model.ID = id

	var err error
	model, err = t2f.saveUC.Save(ctx, model)
	if err != nil {
		return nil, err
	}

	return mapper.MapTest2ModelToDTO(model), nil
}

func (t2f *Test2) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errs.NewInvalidArgumentError("id", "required")
	}

	return t2f.deleteUC.Delete(ctx, id)
}
