package interceptors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type mockAuthenticator struct {
	mock.Mock
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, credentials any) (*auth.Subject, error) {
	args := m.Called(ctx, credentials)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Subject), args.Error(1)
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestAuthUnaryInterceptor(t *testing.T) {
	userSubj := auth.NewSubject("u-1", "ivan", auth.SubjectUser, []string{"admin"}, nil)

	t.Run("success authentication injects subject into context", func(t *testing.T) {
		authMock := new(mockAuthenticator)
		authMock.On("Authenticate", mock.Anything, "valid-creds").Return(userSubj, nil).Once()

		interceptor := AuthUnaryInterceptor([]GRPCAuthProvider{
			{
				Extract:       func(_ context.Context) (any, error) { return "valid-creds", nil },
				Authenticator: authMock,
			},
		}, true, nil)

		var gotSubj *auth.Subject
		resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, _ any) (any, error) {
			gotSubj, _ = auth.FromContext(ctx)
			return "ok", nil
		})

		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.Equal(t, userSubj, gotSubj)
		authMock.AssertExpectations(t)
	})

	t.Run("failed authentication returns Unauthenticated status", func(t *testing.T) {
		authMock := new(mockAuthenticator)
		authMock.On("Authenticate", mock.Anything, "bad-creds").Return(nil, errors.New("bad token")).Once()

		interceptor := AuthUnaryInterceptor([]GRPCAuthProvider{
			{
				Extract:       func(_ context.Context) (any, error) { return "bad-creds", nil },
				Authenticator: authMock,
			},
		}, false, nil)

		resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permissive mode without credentials injects auth.Guest", func(t *testing.T) {
		interceptor := AuthUnaryInterceptor(nil, false, nil)

		var gotSubj *auth.Subject
		_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, _ any) (any, error) {
			gotSubj, _ = auth.FromContext(ctx)
			return "ok", nil
		})

		require.NoError(t, err)
		assert.Equal(t, auth.Guest, gotSubj)
	})

	t.Run("strict mode without credentials returns Unauthenticated", func(t *testing.T) {
		interceptor := AuthUnaryInterceptor(nil, true, nil)

		_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		})

		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestAuthStreamInterceptor(t *testing.T) {
	userSubj := auth.NewSubject("u-1", "ivan", auth.SubjectUser, []string{"admin"}, nil)

	authMock := new(mockAuthenticator)
	authMock.On("Authenticate", mock.Anything, "stream-creds").Return(userSubj, nil).Once()

	interceptor := AuthStreamInterceptor([]GRPCAuthProvider{
		{
			Extract:       func(_ context.Context) (any, error) { return "stream-creds", nil },
			Authenticator: authMock,
		},
	}, true, nil)

	stream := &mockServerStream{ctx: context.Background()}
	var gotSubj *auth.Subject

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, func(_ any, ss grpc.ServerStream) error {
		gotSubj, _ = auth.FromContext(ss.Context())
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, userSubj, gotSubj)
	authMock.AssertExpectations(t)
}

func TestGRPCExtractors(t *testing.T) {
	t.Run("NewGRPCBasicExtractor", func(t *testing.T) {
		extract := NewGRPCBasicExtractor("")

		_, err := extract(context.Background())
		assert.ErrorIs(t, err, ErrCredentialsNotFound)

		encoded := base64.StdEncoding.EncodeToString([]byte("svc-user:svc-pass"))
		md := metadata.Pairs("authorization", "Basic "+encoded)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		creds, err := extract(ctx)
		require.NoError(t, err)
		assert.Equal(t, &auth.BasicCredentials{Username: "svc-user", Password: "svc-pass"}, creds)
	})

	t.Run("NewGRPCCertExtractor", func(t *testing.T) {
		extract := NewGRPCCertExtractor()

		_, err := extract(context.Background())
		assert.ErrorIs(t, err, ErrCredentialsNotFound)

		cert := &x509.Certificate{Subject: pkix.Name{CommonName: "audit-client"}}
		p := &peer.Peer{
			AuthInfo: credentials.TLSInfo{
				State: tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{cert},
				},
			},
		}
		ctx := peer.NewContext(context.Background(), p)

		creds, err := extract(ctx)
		require.NoError(t, err)
		assert.Equal(t, &auth.CertCredentials{Certificate: cert}, creds)
	})
}
