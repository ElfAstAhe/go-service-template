package jwt

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestGRPCExtractor_TokenFromMeta(t *testing.T) {
	expectedToken := &jwt.Token{Valid: true}

	tests := []struct {
		name           string
		md             metadata.MD
		metaName       string
		setupMock      func(m *mockTokenExtractor)
		expectErr      bool
		expectedErrTyp any
	}{
		{
			name:     "success extract token with Bearer prefix",
			md:       metadata.Pairs("authorization", "Bearer   valid.grpc.token  "),
			metaName: "authorization",
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "valid.grpc.token").Return(expectedToken, nil).Once()
			},
			expectErr: false,
		},
		{
			name:     "success extract raw token without prefix (fallback)",
			md:       metadata.Pairs("authorization", "raw.grpc.token"),
			metaName: "Authorization",
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "raw.grpc.token").Return(expectedToken, nil).Once()
			},
			expectErr: false,
		},
		{
			name:           "failed - empty metadata key name",
			md:             metadata.Pairs("authorization", "Bearer valid.grpc.token"),
			metaName:       "   ",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:           "failed - nil metadata",
			md:             nil,
			metaName:       "authorization",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:           "failed - missing key in metadata",
			md:             metadata.Pairs("other-key", "value"),
			metaName:       "authorization",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:           "failed - empty value in metadata",
			md:             metadata.Pairs("authorization", "   "),
			metaName:       "authorization",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:     "failed - underlying tokenExtractor returns error",
			md:       metadata.Pairs("authorization", "Bearer invalid.grpc.token"),
			metaName: "authorization",
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "invalid.grpc.token").Return(nil, errors.New("invalid signature")).Once()
			},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTokenExtractor)
			tt.setupMock(m)

			ge := NewGRPCExtractor(m, DefaultGRPCTokenPrefix)
			token, err := ge.TokenFromMeta(tt.md, tt.metaName)

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

func TestGRPCExtractor_TokenFromContext(t *testing.T) {
	expectedToken := &jwt.Token{Valid: true}

	tests := []struct {
		name           string
		ctx            context.Context
		metaName       string
		setupMock      func(m *mockTokenExtractor)
		expectErr      bool
		expectedErrTyp any
	}{
		{
			name:     "success extract token from incoming gRPC context",
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer ctx.jwt.token")),
			metaName: "authorization",
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "ctx.jwt.token").Return(expectedToken, nil).Once()
			},
			expectErr: false,
		},
		{
			name:           "failed - empty metadata key name",
			ctx:            metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer ctx.jwt.token")),
			metaName:       "",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:           "failed - nil context",
			ctx:            nil,
			metaName:       "authorization",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:           "failed - context without incoming metadata",
			ctx:            context.Background(),
			metaName:       "authorization",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:           "failed - incoming metadata without target key",
			ctx:            metadata.NewIncomingContext(context.Background(), metadata.Pairs("trace-id", "123")),
			metaName:       "authorization",
			setupMock:      func(_ *mockTokenExtractor) {},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:     "failed - underlying tokenExtractor returns error",
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer bad.ctx.token")),
			metaName: "authorization",
			setupMock: func(m *mockTokenExtractor) {
				m.On("Token", "bad.ctx.token").Return(nil, errors.New("token expired")).Once()
			},
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTokenExtractor)
			tt.setupMock(m)

			ge := NewGRPCExtractor(m, DefaultGRPCTokenPrefix)
			token, err := ge.TokenFromContext(tt.ctx, tt.metaName)

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
