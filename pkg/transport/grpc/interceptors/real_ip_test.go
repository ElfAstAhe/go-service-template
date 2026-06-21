package interceptors

import (
	"context"
	"net"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Тестовый сетевой адрес для реализации net.Addr
type testAddr struct {
	net.Addr
	str string
}

func (a testAddr) String() string { return a.str }

func TestRealIPExtractorUSInterceptor_UnaryServerInterceptor(t *testing.T) {
	// Хелпер для создания контекста с peer (физическим сокетом соединения)
	withPeer := func(ctx context.Context, addr string) context.Context {
		return peer.NewContext(ctx, &peer.Peer{
			Addr: testAddr{str: addr},
		})
	}

	// Хелпер для создания контекста с входящими gRPC-метаданными
	withMetadata := func(ctx context.Context, md map[string]string) context.Context {
		return metadata.NewIncomingContext(ctx, metadata.New(md))
	}

	tests := []struct {
		name           string
		allowedHeaders []string
		setupCtx       func(ctx context.Context) context.Context
		wantIP         string
	}{
		{
			name:           "Дефолтный интерцептор: берет первый IP из x-forwarded-for",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"x-forwarded-for": "203.0.113.195, 70.41.3.18",
				})
			},
			wantIP: "203.0.113.195",
		},
		{
			name:           "Поддержка префикса grpcgateway- для x-forwarded-for",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"grpcgateway-x-forwarded-for": "203.0.113.195",
				})
			},
			wantIP: "203.0.113.195",
		},
		{
			name:           "Дефолтный интерцептор: фолбек на x-real-ip, если xff пустой",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"x-real-ip": "198.51.100.1",
				})
			},
			wantIP: "198.51.100.1",
		},
		{
			name:           "Защита от спуфинга: игнорирует cf-connecting-ip, если он не разрешен явно",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"cf-connecting-ip": "1.1.1.1",
					"x-real-ip":        "198.51.100.1",
				})
			},
			wantIP: "198.51.100.1",
		},
		{
			name:           "Явное включение CDN: приоритет cf-connecting-ip над xff",
			allowedHeaders: []string{"CF-Connecting-IP", "X-Forwarded-For"},
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"cf-connecting-ip": "1.1.1.1",
					"x-forwarded-for":  "203.0.113.195",
				})
			},
			wantIP: "1.1.1.1",
		},
		{
			name:           "Валидация: игнорирует некорректный IP в метаданных и идет ниже по каскаду",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"x-forwarded-for": "invalid-ip-string",
					"x-real-ip":       "198.51.100.1",
				})
			},
			wantIP: "198.51.100.1",
		},
		{
			name:           "Кастомный заголовок: корректная обработка",
			allowedHeaders: []string{"X-Custom-Private-IP"},
			setupCtx: func(ctx context.Context) context.Context {
				ctx = withPeer(ctx, "192.168.1.1:12345")
				return withMetadata(ctx, map[string]string{
					"x-custom-private-ip": "10.0.0.5",
				})
			},
			wantIP: "10.0.0.5",
		},
		{
			name:           "Фолбек на Peer: когда все метаданные отсутствуют",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				return withPeer(ctx, "192.0.2.1:54321")
			},
			wantIP: "192.0.2.1",
		},
		{
			name:           "Фолбек на Peer без порта (например, unix-сокет)",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				return withPeer(ctx, "192.0.2.2")
			},
			wantIP: "192.0.2.2",
		},
		{
			name:           "Поддержка IPv6 в x-forwarded-for",
			allowedHeaders: defaultHeaders,
			setupCtx: func(ctx context.Context) context.Context {
				return withMetadata(ctx, map[string]string{
					"x-forwarded-for": "2001:db8::1, 192.0.2.1",
				})
			},
			wantIP: "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewRealIPExtractorUSInterceptor(tt.allowedHeaders...)
			interceptor := extractor.UnaryServerInterceptor()

			ctx := context.Background()
			if tt.setupCtx != nil {
				ctx = tt.setupCtx(ctx)
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/TestMethod",
			}

			handler := func(handlerCtx context.Context, req any) (any, error) {
				// Используем обновленный метод transport.RealIP(ctx)
				gotIP := transport.RealIP(handlerCtx)
				if gotIP != tt.wantIP {
					t.Errorf("transport.RealIP() = %q, ожидался %q", gotIP, tt.wantIP)
				}
				return "response", nil
			}

			_, err := interceptor(ctx, "request", info, handler)
			if err != nil {
				t.Fatalf("Интерцептор вернул непредвиденную ошибку: %v", err)
			}
		})
	}
}

func TestRealIPExtractorUSInterceptor_EmptyHeaders(t *testing.T) {
	extractor := NewRealIPExtractorUSInterceptor()
	interceptor := extractor.UnaryServerInterceptor()

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/TestMethod"}

	handlerCalled := false
	handler := func(handlerCtx context.Context, req any) (any, error) {
		handlerCalled = true
		return "response", nil
	}

	_, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("Ошибка: %v", err)
	}
	if !handlerCalled {
		t.Error("Следующий gRPC хендлер не был вызван")
	}
}
