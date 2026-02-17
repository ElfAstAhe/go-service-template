package domain

type Identity[ID any] interface {
	GetID() ID
	SetID(id ID)
	Validate() error
}
