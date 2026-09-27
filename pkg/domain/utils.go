package domain

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/google/uuid"
)

// EntitiesToIDList трансформирует срез доменных сущностей в плоский срез их уникальных идентификаторов ID.
// За счет пре-аллокации памяти точного размера работает с максимальной скоростью без лишней нагрузки на GC.
func EntitiesToIDList[ID comparable, T Entity[ID]](src []T) []ID {
	res := make([]ID, len(src))

	for index, entity := range src {
		res[index] = entity.GetID()
	}

	return res
}

// AssignUUIDv7 генерирует криптографически стойкий, упорядоченный по времени идентификатор UUIDv7
// и атомарно инжектирует его в строковое поле ID доменной сущности.
// 💡 Рекомендуется для первичных ключей СУБД (PostgreSQL) во избежание деградации B-Tree индексов.
func AssignUUIDv7[T Entity[string]](entity T) error {
	newID, err := uuid.NewV7()
	if err != nil {
		return errs.NewBllError("AssignUUIDv7", "generate new id", err)
	}

	entity.SetID(newID.String())

	return nil
}

// AssignUUIDv4 генерирует случайный идентификатор UUIDv4 (через криптографически стойкий метод NewRandom)
// и атомарно инжектирует его в строковое поле ID доменной сущности.
func AssignUUIDv4[T Entity[string]](entity T) error {
	newID, err := uuid.NewRandom()
	if err != nil {
		// ИСПРАВЛЕНО: Литерал операции теперь корректно указывает на текущий метод AssignUUIDv4
		return errs.NewBllError("AssignUUIDv4", "generate new id", err)
	}

	entity.SetID(newID.String())

	return nil
}

// AssignUUIDv1 генерирует классический идентификатор UUIDv1 (на основе времени и MAC-адреса хоста)
// и атомарно инжектирует его в строковое поле ID доменной сущности.
func AssignUUIDv1[T Entity[string]](entity T) error {
	newID, err := uuid.NewUUID()
	if err != nil {
		return errs.NewBllError("AssignUUIDv1", "generate new id", err)
	}

	entity.SetID(newID.String())

	return nil
}
