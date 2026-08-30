package domain

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/google/uuid"
)

func EntitiesToIDList[T Entity[ID], ID comparable](src []T) []ID {
	res := make([]ID, len(src))
	if len(res) == 0 {
		return res
	}
	for index, entity := range src {
		res[index] = entity.GetID()
	}

	return res
}

func AssignUUIDv7[T Entity[string]](entity T) error {
	newID, err := uuid.NewV7()
	if err != nil {
		return errs.NewBllError("AssignUUIDv7", "generate new id", err)
	}

	entity.SetID(newID.String())

	return nil
}
