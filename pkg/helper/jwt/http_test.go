package jwt

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockTokenExtractor симулирует контракт TokenExtractor (аналогично коду, генерируемому mockery).
type mockTokenExtractor struct {
	mock.Mock
}

func (m *mockTokenExtractor) Claims(token *jwt.Token) (*SubjectClaims, error) {
	args := m.Called(token)
	var res *SubjectClaims
	if args.Get(0) != nil {
		res = args.Get(0).(*SubjectClaims)
	}
	return res, args.Error(1)
}

func (m *mockTokenExtractor) Token(tokenStr string) (*jwt.Token, error) {
	args := m.Called(tokenStr)
	var res *jwt.Token
	if args.Get(0) != nil {
		res = args.Get(0).(*jwt.Token)
	}
	return res, args.Error(1)
}

func TestHTTPExtractor_TokenFromRequestHeader(t *testing.T) {
	expectedToken := &jwt.Token{Valid: true}

	tests := []struct {
		name           string
		headerName     string
		setupReq       func() *http.Request
		setupMock      func(m *mockTokenExtractor)
		expectErr      bool
		expectedErrTyp any
	}{
		{
			name:       "success extract token with Bearer prefix and extra spaces",
			headerName: "Authorization",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer   valid.jwt.token  ")
				return req
			},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "valid.jwt.token").Return(expectedToken, nil).Once()
			},
			expectErr: false,
		},
		{
			name:       "failed - empty header name",
			headerName: "   ",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:       "failed - nil request",
			headerName: "Authorization",
			setupReq: func() *http.Request {
				return nil
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:       "failed - missing header value",
			headerName: "Authorization",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:       "failed - header without expected prefix passes empty string to extractor",
			headerName: "Authorization",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
				return req
			},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "").Return(nil, errs.NewInvalidArgumentError("token", "is required")).Once()
			},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:       "failed - underlying tokenExtractor returns error",
			headerName: "Authorization",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer invalid.token")
				return req
			},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "invalid.token").Return(nil, errors.New("signature invalid")).Once()
			},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTokenExtractor)
			tt.setupMock(m)

			he := NewHTTPExtractor(m, DefaultHTTPTokenPrefix)
			token, err := he.TokenFromRequestHeader(tt.setupReq(), tt.headerName)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorAs(t, err, tt.expectedErrTyp)
				assert.Nil(t, token)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedToken, token)
			}

			m.AssertExpectations(t)
		})
	}
}

//goland:noinspection DuplicatedCode
func TestHTTPExtractor_TokenFromRequestCookie(t *testing.T) {
	expectedToken := &jwt.Token{Valid: true}

	tests := []struct {
		name           string
		cookieName     string
		setupReq       func() *http.Request
		setupMock      func(m *mockTokenExtractor)
		expectErr      bool
		expectedErrTyp any
	}{
		{
			name:       "success extract raw token from cookie without prefix",
			cookieName: "access_token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "raw.jwt.token",
				})
				return req
			},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "raw.jwt.token").Return(expectedToken, nil).Once()
			},
			expectErr: false,
		},
		{
			name:       "success extract token from cookie with Bearer prefix",
			cookieName: "access_token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "Bearer prefixed.jwt.token",
				})
				return req
			},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "prefixed.jwt.token").Return(expectedToken, nil).Once()
			},
			expectErr: false,
		},
		{
			name:       "failed - empty cookie name",
			cookieName: "",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:       "failed - nil request",
			cookieName: "access_token",
			setupReq: func() *http.Request {
				return nil
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:       "failed - cookie not found in request",
			cookieName: "access_token",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:       "failed - cookie value is empty",
			cookieName: "access_token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "",
				})
				return req
			},
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:       "failed - underlying tokenExtractor returns error",
			cookieName: "access_token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "bad.jwt.token",
				})
				return req
			},
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "bad.jwt.token").Return(nil, errors.New("parse error")).Once()
			},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTokenExtractor)
			tt.setupMock(m)

			he := NewHTTPExtractor(m, DefaultHTTPTokenPrefix)
			token, err := he.TokenFromRequestCookie(tt.setupReq(), tt.cookieName)

			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorAs(t, err, tt.expectedErrTyp)
				assert.Nil(t, token)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedToken, token)
			}

			m.AssertExpectations(t)
		})
	}
}
