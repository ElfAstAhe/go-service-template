package jwt

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubjectClaims(t *testing.T) {
	t.Run("success create empty instance", func(t *testing.T) {
		claims := newSubjectClaims()

		require.NotNil(t, claims)
		assert.Empty(t, claims.Subject)
		assert.Empty(t, claims.SubjectID)
		assert.Empty(t, claims.SubjectType)
		assert.False(t, claims.Admin)
		assert.Nil(t, claims.Roles)
	})
}

func TestSubjectClaims_ValidateSubject(t *testing.T) {
	tests := []struct {
		name        string
		claims      *SubjectClaims
		expectErr   bool
		expectedMsg string
	}{
		{
			name: "success validation",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "ivan_prod",
				},
				SubjectID:   "user-uuid-1234-5678",
				SubjectType: "user",
			},
			expectErr: false,
		},
		{
			name: "failed - empty subject",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "",
				},
				SubjectID:   "user-uuid-1234-5678",
				SubjectType: "user",
			},
			expectErr:   true,
			expectedMsg: "subject required",
		},
		{
			name: "failed - space-only subject",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "   ",
				},
				SubjectID:   "user-uuid-1234-5678",
				SubjectType: "user",
			},
			expectErr:   true,
			expectedMsg: "subject required",
		},
		{
			name: "failed - empty subject type",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "ivan_prod",
				},
				SubjectID:   "user-uuid-1234-5678",
				SubjectType: "",
			},
			expectErr:   true,
			expectedMsg: "subject type required",
		},
		{
			name: "failed - space-only subject type",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "ivan_prod",
				},
				SubjectID:   "user-uuid-1234-5678",
				SubjectType: "  \t  ",
			},
			expectErr:   true,
			expectedMsg: "subject type required",
		},
		{
			name: "failed - empty subject id",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "ivan_prod",
				},
				SubjectID:   "",
				SubjectType: "user",
			},
			expectErr:   true,
			expectedMsg: "subject id required",
		},
		{
			name: "failed - space-only subject id",
			claims: &SubjectClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "ivan_prod",
				},
				SubjectID:   "     ",
				SubjectType: "user",
			},
			expectErr:   true,
			expectedMsg: "subject id required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.ValidateSubject()

			if tt.expectErr {
				var jwtErr *errs.UtlJWTError
				require.Error(t, err)
				require.ErrorAs(t, err, &jwtErr)
				assert.Contains(t, err.Error(), tt.expectedMsg)
				return
			}

			assert.NoError(t, err)
		})
	}
}
