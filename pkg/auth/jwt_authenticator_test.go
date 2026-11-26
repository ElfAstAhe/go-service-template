package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkghlpjwt "github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTokenExtractor struct {
	mock.Mock
}

func (m *mockTokenExtractor) Claims(token *jwt.Token) (*pkghlpjwt.SubjectClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pkghlpjwt.SubjectClaims), args.Error(1)
}

func (m *mockTokenExtractor) Token(tokenStr string) (*jwt.Token, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func TestNewJWTAuthenticator(t *testing.T) {
	t.Run("nil interface extractor returns InvalidArgumentError", func(t *testing.T) {
		authenticator, err := NewJWTAuthenticator(nil)

		require.Error(t, err)
		assert.Nil(t, authenticator)

		var invErr *errs.InvalidArgumentError
		assert.ErrorAs(t, err, &invErr)
	})

	t.Run("typed nil extractor returns InvalidArgumentError", func(t *testing.T) {
		var nilExt *mockTokenExtractor

		authenticator, err := NewJWTAuthenticator(nilExt)

		require.Error(t, err)
		assert.Nil(t, authenticator)

		var invErr *errs.InvalidArgumentError
		assert.ErrorAs(t, err, &invErr)
	})

	t.Run("valid extractor creates instance", func(t *testing.T) {
		ext := new(mockTokenExtractor)

		authenticator, err := NewJWTAuthenticator(ext)

		require.NoError(t, err)
		require.NotNil(t, authenticator)
	})
}

func TestJWTAuthenticator_Authenticate(t *testing.T) {
	dummyToken := &jwt.Token{Raw: "header.payload.signature"}
	extractorErr := errors.New("broken signature")
	var typedNilCreds *JWTCredentials

	tests := []struct {
		name           string
		credentials    any
		setupMock      func(m *mockTokenExtractor)
		wantErrType    any
		wantErrCause   error
		wantErrContain string
		wantSubject    *Subject
	}{
		{
			name:        "untyped nil credentials returns InvalidArgumentError",
			credentials: nil,
			setupMock:   func(_ *mockTokenExtractor) {},
			wantErrType: new(*errs.InvalidArgumentError),
		},
		{
			name:        "typed nil JWTCredentials returns InvalidArgumentError",
			credentials: typedNilCreds,
			setupMock:   func(_ *mockTokenExtractor) {},
			wantErrType: new(*errs.InvalidArgumentError),
		},
		{
			name:           "wrong credentials type returns UtlAuthError",
			credentials:    &BasicCredentials{Username: "user", Password: "pwd"},
			setupMock:      func(_ *mockTokenExtractor) {},
			wantErrType:    new(*errs.UtlAuthError),
			wantErrContain: "invalid credentials type",
		},
		{
			name:           "nil Token inside JWTCredentials returns UtlAuthError",
			credentials:    &JWTCredentials{Token: nil},
			setupMock:      func(_ *mockTokenExtractor) {},
			wantErrType:    new(*errs.UtlAuthError),
			wantErrContain: "empty JWT token",
		},
		{
			name:        "extractor Claims error returns wrapped UtlAuthError",
			credentials: &JWTCredentials{Token: dummyToken},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Claims", dummyToken).Return(nil, extractorErr).Once()
			},
			wantErrType:    new(*errs.UtlAuthError),
			wantErrCause:   extractorErr,
			wantErrContain: "jwt claims extraction failed",
		},
		// ToDo: fix or remove
		//{
		//	name:        "invalid claims mapping returns UtlAuthError",
		//	credentials: &JWTCredentials{Token: dummyToken},
		//	setupMock: func(m *mockTokenExtractor) {
		//		m.On("Claims", dummyToken).Return(&pkghlpjwt.SubjectClaims{
		//			RegisteredClaims: jwt.RegisteredClaims{
		//				Subject: "", // пустой Subject вызовет ошибку маппинга
		//			},
		//			SubjectType: string(SubjectUser),
		//		}, nil).Once()
		//	},
		//	wantErrType:    new(*errs.UtlAuthError),
		//	wantErrContain: "failed to map jwt claims to subject",
		//},
		{
			name:        "valid claims return authenticated Subject",
			credentials: &JWTCredentials{Token: dummyToken},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Claims", dummyToken).Return(&pkghlpjwt.SubjectClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "ivan",
					},
					SubjectID:   "user-42",
					SubjectType: string(SubjectUser),
					Roles:       []string{"admin", "developer"},
				}, nil).Once()
			},
			wantSubject: &Subject{
				ID:    "user-42",
				Name:  "ivan",
				Type:  SubjectUser,
				Roles: map[string]struct{}{"admin": {}, "developer": {}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := new(mockTokenExtractor)
			tt.setupMock(ext)

			authenticator, err := NewJWTAuthenticator(ext)
			require.NoError(t, err)

			subj, err := authenticator.Authenticate(context.Background(), tt.credentials)

			if tt.wantErrType != nil {
				require.Error(t, err)
				assert.Nil(t, subj)
				assert.ErrorAs(t, err, tt.wantErrType)

				if tt.wantErrCause != nil {
					assert.ErrorIs(t, err, tt.wantErrCause)
				}
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}

				ext.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, subj)
			assert.Equal(t, tt.wantSubject.ID, subj.ID)
			assert.Equal(t, tt.wantSubject.Name, subj.Name)
			assert.Equal(t, tt.wantSubject.Type, subj.Type)
			assert.Equal(t, tt.wantSubject.Roles, subj.Roles)
			assert.True(t, subj.IsAuthenticated())
			ext.AssertExpectations(t)
		})
	}
}
