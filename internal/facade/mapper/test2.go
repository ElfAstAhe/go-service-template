package mapper

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

func MapModelToDTO(instance *domain.Test2) *dto.Test2 {
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

func MapDTOToModel(instance *dto.Test2) *domain.Test2 {
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
