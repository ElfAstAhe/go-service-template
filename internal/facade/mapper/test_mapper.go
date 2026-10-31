package mapper

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

func MapTestDtoToModel(testDTO *dto.Test) *domain.Test {
	if utils.IsNil(testDTO) {
		return nil
	}

	res := domain.NewEmptyTest()

	res.ID = testDTO.ID
	res.Code = testDTO.Code
	res.Name = testDTO.Name
	res.Description = testDTO.Description
	res.CreatedAt = testDTO.RegisteredAt
	res.ModifiedAt = testDTO.UpdatedAt

	return res
}

func MapTestModelToDto(model *domain.Test) *dto.Test {
	if utils.IsNil(model) {
		return nil
	}

	res := &dto.Test{}
	res.ID = model.ID
	res.Code = model.Code
	res.Name = model.Name
	res.Description = model.Description
	res.RegisteredAt = model.CreatedAt
	res.UpdatedAt = model.ModifiedAt

	return res
}

func MapTestModelsToDtos(models []*domain.Test) []*dto.Test {
	if len(models) == 0 {
		return make([]*dto.Test, 0)
	}

	res := make([]*dto.Test, len(models))

	for i, model := range models {
		res[i] = MapTestModelToDto(model)
	}

	return res
}
