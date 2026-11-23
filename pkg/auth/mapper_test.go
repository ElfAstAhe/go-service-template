package auth

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapClaimsToSubject(t *testing.T) {
	tests := []struct {
		name           string
		claims         *jwt.SubjectClaims
		expectErr      bool
		expectedErrTyp any
		assertSubject  func(t *testing.T, s *Subject)
	}{
		{
			name: "success map user claims with roles",
			claims: jwt.NewSubjectClaimsBuilder().
				WithSubjectID("user-uuid-1").
				WithSubject("ivan").
				WithSubjectType("user").
				WithRoles("audit:read", "audit:write").
				Build(),
			expectErr: false,
			assertSubject: func(t *testing.T, s *Subject) {
				assert.Equal(t, "user-uuid-1", s.ID)
				assert.Equal(t, "ivan", s.Name)
				assert.Equal(t, SubjectUser, s.Type)
				assert.True(t, s.HasRole("audit:read"))
				assert.True(t, s.HasRole("audit:write"))
				assert.False(t, s.HasRole("admin"))
				assert.Empty(t, s.Metadata)
			},
		},
		{
			name: "success map guest claims without roles",
			claims: jwt.NewSubjectClaimsBuilder().
				WithSubjectID("").
				WithSubject("anonymous").
				WithSubjectType("guest").
				Build(),
			expectErr: false,
			assertSubject: func(t *testing.T, s *Subject) {
				assert.Empty(t, s.ID)
				assert.Equal(t, "anonymous", s.Name)
				assert.Equal(t, SubjectGuest, s.Type)
				assert.Empty(t, s.Roles)
				assert.True(t, s.IsGuest())
			},
		},
		{
			name:           "failed - nil claims",
			claims:         nil,
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name: "failed - unknown subject type",
			claims: jwt.NewSubjectClaimsBuilder().
				WithSubjectID("uuid-1").
				WithSubject("bot").
				WithSubjectType("unknown_type").
				Build(),
			expectErr:      true,
			expectedErrTyp: new(*errs.CommonError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, err := MapClaimsToSubject(tt.claims)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorAs(t, err, tt.expectedErrTyp)
				assert.Nil(t, subject)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, subject)
			tt.assertSubject(t, subject)
		})
	}
}

func TestMapSubjectToClaimsBuilder(t *testing.T) {
	t.Run("success map user subject to claims builder", func(t *testing.T) {
		subject := NewSubject("srv-uuid-99", "billing-worker", SubjectService, []string{"m2m:events", "m2m:metrics"}, nil)

		builder, err := MapSubjectToClaimsBuilder(subject)

		require.NoError(t, err)
		require.NotNil(t, builder)

		claims := builder.Build()
		assert.Equal(t, "srv-uuid-99", claims.SubjectID)
		assert.Equal(t, "billing-worker", claims.Subject)
		assert.Equal(t, "service", claims.SubjectType)
		assert.ElementsMatch(t, []string{"m2m:events", "m2m:metrics"}, claims.Roles)
	})

	t.Run("success map guest subject with empty roles", func(t *testing.T) {
		builder, err := MapSubjectToClaimsBuilder(Guest)

		require.NoError(t, err)
		require.NotNil(t, builder)

		claims := builder.Build()
		assert.Empty(t, claims.SubjectID)
		assert.Equal(t, "Guest", claims.Subject)
		assert.Equal(t, "guest", claims.SubjectType)
		assert.Empty(t, claims.Roles)
	})

	t.Run("failed - nil subject", func(t *testing.T) {
		builder, err := MapSubjectToClaimsBuilder(nil)

		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
		assert.Nil(t, builder)
	})
}
