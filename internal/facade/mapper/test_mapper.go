package mapper

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
)

func TestDtoToModel(testDTO *dto.TestDTO) *domain.Test {
	if testDTO == nil {
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

func TestModelToDto(model *domain.Test) *dto.TestDTO {
	if model == nil {
		return nil
	}

	res := &dto.TestDTO{}
	res.ID = model.ID
	res.Code = model.Code
	res.Name = model.Name
	res.Description = model.Description
	res.RegisteredAt = model.CreatedAt
	res.UpdatedAt = model.ModifiedAt

	return res
}

func TestModelsToDtos(models []*domain.Test) []*dto.TestDTO {
	if len(models) == 0 {
		return make([]*dto.TestDTO, 0)
	}

	res := make([]*dto.TestDTO, len(models))

	for i, model := range models {
		res[i] = TestModelToDto(model)
	}

	return res
}
