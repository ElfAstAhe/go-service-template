package auth

import (
	"context"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSubjectAndFromContext(t *testing.T) {
	validSubject := NewSubject("uuid-123", "ivan", SubjectUser, []string{"admin"}, nil)

	t.Run("success store and extract subject", func(t *testing.T) {
		ctx := WithSubject(context.Background(), validSubject)

		extracted, err := FromContext(ctx)

		require.NoError(t, err)
		require.NotNil(t, extracted)
		assert.Equal(t, validSubject, extracted)
	})

	t.Run("success store and extract Guest singleton", func(t *testing.T) {
		ctx := WithSubject(context.Background(), Guest)

		extracted, err := FromContext(ctx)

		require.NoError(t, err)
		require.NotNil(t, extracted)
		assert.Equal(t, Guest, extracted)
		assert.True(t, extracted.IsGuest())
		assert.False(t, extracted.IsAuthenticated())
	})

	t.Run("failed - empty background context", func(t *testing.T) {
		extracted, err := FromContext(context.Background())

		var authErr *errs.UtlAuthError
		require.Error(t, err)
		assert.ErrorAs(t, err, &authErr)
		assert.Nil(t, extracted)
	})

	t.Run("failed - nil subject stored in context", func(t *testing.T) {
		ctx := WithSubject(context.Background(), nil)

		extracted, err := FromContext(ctx)

		var authErr *errs.UtlAuthError
		require.Error(t, err)
		assert.ErrorAs(t, err, &authErr)
		assert.Nil(t, extracted)
	})

	t.Run("failed - collision protection with foreign key type", func(t *testing.T) {
		type foreignKey struct{}
		ctx := context.WithValue(context.Background(), foreignKey{}, validSubject)

		extracted, err := FromContext(ctx)

		var authErr *errs.UtlAuthError
		require.Error(t, err)
		assert.ErrorAs(t, err, &authErr)
		assert.Nil(t, extracted)
	})
}
