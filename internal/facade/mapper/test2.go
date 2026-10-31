package mapper

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

func MapTest2ModelToDTO(instance *domain.Test2) *dto.Test2 {
	if utils.IsNil(instance) {
		return nil
	}

	return &dto.Test2{
		ID:         instance.ID,
		Code:       instance.Code,
		Name:       instance.Name,
		CreatedAt:  instance.CreatedAt,
		ModifiedAt: instance.ModifiedAt,
	}
}

func MapTest2DTOToModel(instance *dto.Test2) *domain.Test2 {
	if utils.IsNil(instance) {
		return nil
	}

	res := domain.NewEmptyTest2()
	res.ID = instance.ID
	res.Code = instance.Code
	res.Name = instance.Name
	res.Description = instance.Description
	res.CreatedAt = instance.CreatedAt
	res.ModifiedAt = instance.ModifiedAt

	return res
}

func MapTest2ModelsToDTO(models []*domain.Test2) []*dto.Test2 {
	if len(models) == 0 {
		return []*dto.Test2{}
	}
	res := make([]*dto.Test2, 0, len(models))
	for _, model := range models {
		res = append(res, MapTest2ModelToDTO(model))
	}

	return res
}
