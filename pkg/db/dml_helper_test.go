package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNames(t *testing.T) {
	t.Run("BuildPKConstraintName", func(t *testing.T) {
		assert.Equal(t, "users_pk", BuildPKConstraintName("users"))
	})

	t.Run("BuildUKConstraintName_SingleField", func(t *testing.T) {
		// Для одного поля должен вернуть само имя без хеша
		assert.Equal(t, "users_email_uk", BuildUKConstraintName("users", "email"))
	})

	t.Run("BuildUKConstraintName_MultipleFields", func(t *testing.T) {
		// Для нескольких полей должен быть хеш
		name1 := BuildUKConstraintName("users", "first_name", "last_name")
		name2 := BuildUKConstraintName("users", "first_name", "last_name")

		assert.Contains(t, name1, "users_")
		assert.Contains(t, name1, "_uk")
		assert.Equal(t, name1, name2, "Имена должны быть детерминированы")
	})

	t.Run("BuildUKConstraintName_EmptyFields", func(t *testing.T) {
		assert.Equal(t, "users_uk", BuildUKConstraintName("users"))
	})
}

func TestBuildFieldNamesHash(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", buildFieldNamesHash())
	})

	t.Run("Single", func(t *testing.T) {
		assert.Equal(t, "field1", buildFieldNamesHash("field1"))
	})

	t.Run("Multiple_Consistency", func(t *testing.T) {
		h1 := buildFieldNamesHash("a", "b", "c")
		h2 := buildFieldNamesHash("a", "b", "c")
		assert.Equal(t, h1, h2)
		assert.Len(t, h1, 32, "MD5 хеш должен быть длиной 32 символа")
	})

	t.Run("Order_Matters", func(t *testing.T) {
		// Разный порядок полей должен давать разный хеш (это нормально для индексов)
		h1 := buildFieldNamesHash("field1", "field2")
		h2 := buildFieldNamesHash("field2", "field1")
		assert.NotEqual(t, h1, h2)
	})
}
