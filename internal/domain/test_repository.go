package domain

import (
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

type TestRepository interface {
	repository.BaseRepository[*Test, string]

	FindByCode(code string) (*Test, error)
}
