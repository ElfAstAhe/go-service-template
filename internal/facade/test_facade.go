package facade

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/ElfAstAhe/go-service-template/internal/facade/mapper"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TestFacade interface {
	Get(ctx context.Context, id string) (*dto.TestDTO, error)
	GetByCode(ctx context.Context, code string) (*dto.TestDTO, error)
	List(ctx context.Context, limit, offset int) ([]*dto.TestDTO, error)
	Create(ctx context.Context, test *dto.TestDTO) (*dto.TestDTO, error)
	Change(ctx context.Context, id string, test *dto.TestDTO) (*dto.TestDTO, error)
	Delete(ctx context.Context, id string) error
}

type TestFacadeImpl struct {
	getUC       usecase.TestGetUseCase
	getByCodeUC usecase.TestGetByCodeUseCase
	listUC      usecase.TestListUseCase
	saveUC      usecase.TestSaveUseCase
	deleteUC    usecase.TestDeleteUseCase
}

func NewTestFacade(
	getUC usecase.TestGetUseCase,
	getByCodeUC usecase.TestGetByCodeUseCase,
	listUC usecase.TestListUseCase,
	saveUC usecase.TestSaveUseCase,
	deleteUC usecase.TestDeleteUseCase,
) *TestFacadeImpl {
	return &TestFacadeImpl{
		getUC:       getUC,
		getByCodeUC: getByCodeUC,
		listUC:      listUC,
		saveUC:      saveUC,
		deleteUC:    deleteUC,
	}
}

func (tf *TestFacadeImpl) Get(ctx context.Context, id string) (*dto.TestDTO, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errs.NewInvalidArgumentError("id", "must not be empty")
	}

	model, err := tf.getUC.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.MapTestModelToDto(model), nil
}

func (tf *TestFacadeImpl) GetByCode(ctx context.Context, code string) (*dto.TestDTO, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errs.NewInvalidArgumentError("code", "must not be empty")
	}

	model, err := tf.getByCodeUC.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	return mapper.MapTestModelToDto(model), nil
}

func (tf *TestFacadeImpl) List(ctx context.Context, limit, offset int) ([]*dto.TestDTO, error) {
	if err := tf.validateList(limit, offset); err != nil {
		return nil, err
	}

	models, err := tf.listUC.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return mapper.MapTestModelsToDtos(models), nil
}

func (tf *TestFacadeImpl) validateList(limit, offset int) error {
	if !(limit > 0) {
		return errs.NewInvalidArgumentError("limit", "must be greater than 0")
	}
	if offset < 0 {
		return errs.NewInvalidArgumentError("offset", "must be greater or equal than 0")
	}
	if limit > 1000 {
		return errs.NewInvalidArgumentError("limit", "must be less or equal than 1000")
	}

	return nil
}

func (tf *TestFacadeImpl) Create(ctx context.Context, test *dto.TestDTO) (*dto.TestDTO, error) {
	if test == nil {
		return nil, errs.NewInvalidArgumentError("test", "must not be empty")
	}

	model := mapper.MapTestDtoToModel(test)
	model.ID = ""

	var err error
	model, err = tf.saveUC.Save(ctx, model)
	if err != nil {
		return nil, err
	}

	return mapper.MapTestModelToDto(model), nil
}

func (tf *TestFacadeImpl) Change(ctx context.Context, id string, test *dto.TestDTO) (*dto.TestDTO, error) {
	if test == nil {
		return nil, errs.NewInvalidArgumentError("test", "must not be empty")
	}

	model := mapper.MapTestDtoToModel(test)
	model.ID = id
	var err error
	model, err = tf.saveUC.Save(ctx, model)
	if err != nil {
		return nil, err
	}

	return mapper.MapTestModelToDto(model), nil
}

func (tf *TestFacadeImpl) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errs.NewInvalidArgumentError("id", "must not be empty")
	}

	return tf.deleteUC.Delete(ctx, id)
}
