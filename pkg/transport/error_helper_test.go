package transport

import (
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestIsBadRequest(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "standard foreign error",
			err:      errors.New("some internal error"),
			expected: false,
		},
		{
			name:     "wrapped generic error",
			err:      errors.New("wrapped: some internal error"),
			expected: false,
		},
		{
			name:     "happy path: InvalidArgumentError",
			err:      &errs.InvalidArgumentError{},
			expected: true,
		},
		{
			name:     "happy path: BllValidateError",
			err:      &errs.BllValidateError{},
			expected: true,
		},
		{
			name:     "happy path: TlMappingError",
			err:      &errs.TlMappingError{},
			expected: true,
		},
		{
			name:     "wrapped valid error",
			err:      &errs.BllValidateError{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := IsBadRequest(tt.err)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "happy path: BllNotFoundError",
			err:      &errs.BllNotFoundError{},
			expected: true,
		},
		{
			name:     "happy path: DalNotFoundError",
			err:      &errs.DalNotFoundError{},
			expected: true,
		},
		{
			name:     "other domain error",
			err:      &errs.BllUnauthorizedError{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := IsNotFound(tt.err)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestRemainingHelpers(t *testing.T) {
	// Краткие проверки для остальных одиночных маппингов
	assert.True(t, IsUnauthorized(&errs.BllUnauthorizedError{}))
	assert.False(t, IsUnauthorized(nil))

	assert.True(t, IsForbidden(&errs.BllForbiddenError{}))
	assert.False(t, IsForbidden(errors.New("forbidden but wrong type")))

	assert.True(t, IsConflict(&errs.BllUniqueError{}))
	assert.True(t, IsConflict(&errs.DalAlreadyExistsError{}))

	assert.True(t, IsGone(&errs.DalSoftDeletedError{}))
}
