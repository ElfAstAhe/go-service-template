package mocks

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestEntitiesToIDList(t *testing.T) {
	// prepare
	ent1 := new(MockEntity[string])
	ent1.On("GetID").Return("1")
	ent2 := new(MockEntity[string])
	ent2.On("GetID").Return("2")
	expected := []string{"1", "2"}
	// act
	var res = domain.EntitiesToIDList([]*MockEntity[string]{ent1, ent2})
	// assert
	assert.Equal(t, 2, len(res))
	assert.Equal(t, expected, res)
}
