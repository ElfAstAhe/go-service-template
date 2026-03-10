package domain

type Entity[ID comparable] interface {
	GetID() ID
	SetID(id ID)

	IsExists() bool

	BeforeCreate() error
	BeforeChange() error

	ValidateCreate() error
	ValidateChange() error
}

type SoftDeleteEntity[DEL any] interface {
	GetDeleted() DEL
	SetDeleted(deleted DEL)

	IsDeleted() bool
}
