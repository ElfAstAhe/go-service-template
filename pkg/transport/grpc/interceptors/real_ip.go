package interceptors

import (
	"context"
	"net"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Константы gRPC метаданных (строго в нижнем регистре для HTTP/2)
const (
	HeaderXRealIP                = "x-real-ip"
	HeaderXForwardedFor          = "x-forwarded-for"
	HeaderClientIP               = "client-ip"
	HeaderXClientIP              = "x-client-ip"
	HeaderXClusterClientIP       = "x-cluster-client-ip"
	HeaderCFConnectingIP         = "cf-connecting-ip"
	HeaderTrueClientIP           = "true-client-ip"
	HeaderXCloudDeploymentUserIP = "x-cloud-deployment-user-ip"
	HeaderXAzureClientIP         = "x-azure-clientip"
	HeaderFastlyClientIP         = "fastly-client-ip"
)

var defaultHeaders = []string{
	HeaderXForwardedFor,
	HeaderXRealIP,
}

type RealIPExtractorUSInterceptor struct {
	allowedHeaders map[string]struct{}
	customHeaders  []string
}

// NewRealIPExtractorUSInterceptor принимает список ожидаемых заголовков
func NewRealIPExtractorUSInterceptor(headers ...string) *RealIPExtractorUSInterceptor {
	allowed := make(map[string]struct{}, len(headers))
	var custom []string

	knownHeaders := map[string]struct{}{
		HeaderXRealIP:                {},
		HeaderXForwardedFor:          {},
		HeaderClientIP:               {},
		HeaderXClientIP:              {},
		HeaderXClusterClientIP:       {},
		HeaderCFConnectingIP:         {},
		HeaderTrueClientIP:           {},
		HeaderXCloudDeploymentUserIP: {},
		HeaderXAzureClientIP:         {},
		HeaderFastlyClientIP:         {},
	}

	for _, h := range headers {
		lowered := strings.ToLower(h)
		allowed[lowered] = struct{}{}
		if _, known := knownHeaders[lowered]; !known {
			custom = append(custom, lowered)
		}
	}

	return &RealIPExtractorUSInterceptor{
		allowedHeaders: allowed,
		customHeaders:  custom,
	}
}

func NewDefaultRealIPExtractorUSInterceptor() *RealIPExtractorUSInterceptor {
	return NewRealIPExtractorUSInterceptor(defaultHeaders...)
}

// UnaryServerInterceptor возвращает готовый интерцептор для gRPC сервера
func (re *RealIPExtractorUSInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if len(re.allowedHeaders) == 0 {
			return handler(ctx, req)
		}

		ip := re.extractRemoteIP(ctx)
		return handler(transport.WithRealIP(ctx, ip), req)
	}
}

// Извлечение IP на основе контекста метаданных gRPC
//
//goland:noinspection DuplicatedCode
func (re *RealIPExtractorUSInterceptor) extractRemoteIP(ctx context.Context) string {
	md, mdOk := metadata.FromIncomingContext(ctx)
	if !mdOk {
		return re.fallbackToPeer(ctx)
	}

	// Хелпер для проверки заголовка с учетом стандартного имени и префикса grpcgateway-
	getMetadataValue := func(header string) string {
		if values := md.Get(header); len(values) > 0 && values[0] != "" {
			return values[0]
		}
		if values := md.Get("grpcgateway-" + header); len(values) > 0 && values[0] != "" {
			return values[0]
		}
		return ""
	}

	// 1. Приоритет I: Эксклюзивные заголовки защищенных CDN/Провайдеров облаков
	if _, ok := re.allowedHeaders[HeaderCFConnectingIP]; ok {
		if ip := getMetadataValue(HeaderCFConnectingIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderTrueClientIP]; ok {
		if ip := getMetadataValue(HeaderTrueClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXCloudDeploymentUserIP]; ok {
		if ip := getMetadataValue(HeaderXCloudDeploymentUserIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderFastlyClientIP]; ok {
		if ip := getMetadataValue(HeaderFastlyClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXAzureClientIP]; ok {
		if ip := getMetadataValue(HeaderXAzureClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 2. Приоритет II: Стандартный X-Forwarded-For
	if _, ok := re.allowedHeaders[HeaderXForwardedFor]; ok {
		if xff := getMetadataValue(HeaderXForwardedFor); xff != "" {
			firstIP, _, _ := strings.Cut(xff, ",")
			if cleanIP := re.cleanAndValidateIP(firstIP); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 3. Приоритет III: Очищенный одиночный X-Real-IP
	if _, ok := re.allowedHeaders[HeaderXRealIP]; ok {
		if xrip := getMetadataValue(HeaderXRealIP); xrip != "" {
			if cleanIP := re.cleanAndValidateIP(xrip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 4. Приоритет IV: Альтернативные enterprise-заголовки прокси-систем
	if _, ok := re.allowedHeaders[HeaderClientIP]; ok {
		if ip := getMetadataValue(HeaderClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXClientIP]; ok {
		if ip := getMetadataValue(HeaderXClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXClusterClientIP]; ok {
		if ip := getMetadataValue(HeaderXClusterClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 5. Приоритет V: Кастомная специфика
	for _, header := range re.customHeaders {
		if ip := getMetadataValue(header); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 6. Финальный фолбек на сетевой сокет gRPC соединения
	return re.fallbackToPeer(ctx)
}

// Извлечение IP из физического peer gRPC
func (re *RealIPExtractorUSInterceptor) fallbackToPeer(ctx context.Context) string {
	if pr, ok := peer.FromContext(ctx); ok && pr.Addr != nil {
		host, _, err := net.SplitHostPort(pr.Addr.String())
		if err == nil {
			if cleanIP := re.cleanAndValidateIP(host); cleanIP != "" {
				return cleanIP
			}
		}
		return strings.TrimSpace(pr.Addr.String())
	}
	return ""
}

func (re *RealIPExtractorUSInterceptor) cleanAndValidateIP(ip string) string {
	cleaned := strings.TrimSpace(ip)
	if cleaned == "" {
		return ""
	}
	if net.ParseIP(cleaned) != nil {
		return cleaned
	}
	return ""
}
