package domain

func EntitiesToIDList[T Entity[ID], ID any](src []T) []ID {
	res := make([]ID, len(src))
	if len(res) == 0 {
		return res
	}
	for index, entity := range src {
		res[index] = entity.GetID()
	}

	return res
}
