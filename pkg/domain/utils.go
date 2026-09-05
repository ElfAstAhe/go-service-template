package domain

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/google/uuid"
)

func EntitiesToIDList[ID comparable, T Entity[ID]](src []T) []ID {
	res := make([]ID, len(src))

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

func AssignUUIDv4[T Entity[string]](entity T) error {
	newID, err := uuid.NewUUID()
	if err != nil {
		return errs.NewBllError("AssignUUIDv4", "generate new id", err)
	}

	entity.SetID(newID.String())

	return nil
}
