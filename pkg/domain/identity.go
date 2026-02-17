package domain

type Identity[ID any] interface {
	GetID() ID
	SetID(id ID)

	BeforeCreate() error
	BeforeChange() error

	ValidateCreate() error
	ValidateChange() error
}

type SoftDeleteIdentity[DEL any] interface {
	GetDeleted() DEL
	SetDeleted(deleted DEL)

	IsDeleted() bool
}
