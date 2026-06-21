package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

const (
	// HeaderXRealIP Формат: X-Real-IP: 203.0.113.195. Нюанс: Nginx или Ingress обычно вырезают из него цепочки и прописывают туда ровно один, проверенный IP-адрес предыдущего узла
	HeaderXRealIP string = "X-Real-IP"
	// HeaderXForwardedFor Формат: X-Forwarded-For: client, proxy1, proxy2. Нюанс: Содержит цепочку IP-адресов через запятую. Реальный IP пользователя всегда идет самым первым. Остальные — это цепочка прокси-серверов, через которые пролетел запрос.
	HeaderXForwardedFor    string = "X-Forwarded-For"
	HeaderClientIP         string = "Client-IP"
	HeaderXClientIP        string = "X-Client-IP"
	HeaderXClusterClientIP string = "X-Cluster-Client-IP"
	// HeaderCFConnectingIP Эксклюзив от Cloudflare. Содержит чистый IP-адрес клиента, который подключился к их сети
	HeaderCFConnectingIP string = "CF-Connecting-IP"
	// HeaderTrueClientIP Стандарт для Akamai CDN и некоторых enterprise-настроек Cloudflare
	HeaderTrueClientIP string = "True-Client-IP"
	// HeaderXCloudDeploymentUserIP Используется в Google Cloud Platform (GCP).
	HeaderXCloudDeploymentUserIP string = "X-Cloud-Deployment-User-IP"
	HeaderXAzureClientIP         string = "X-Azure-ClientIP"
	// HeaderFastlyClientIP Используется в Fastly CDN.
	HeaderFastlyClientIP string = "Fastly-Client-IP"
)

// Безопасный дефолтный список. Исключаем CDN-заголовки во избежание спуфинга,
// если приложение развернуто в обычной инфраструктуре.
var defaultHeaders = []string{
	HeaderXForwardedFor,
	HeaderXRealIP,
}

type RealIPExtractor struct {
	// Храним в map для мгновенного поиска O(1) вместо slices.Contains
	allowedHeaders map[string]struct{}
	// Сохраняем исходный порядок кастомных заголовков для шага 5
	customHeaders []string
}

func NewRealIPExtractor(headers ...string) *RealIPExtractor {
	allowed := make(map[string]struct{}, len(headers))
	var custom []string

	// Мапа для проверки принадлежности к "базовым" известным заголовкам
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
		allowed[h] = struct{}{}
		if _, known := knownHeaders[h]; !known {
			custom = append(custom, h)
		}
	}

	return &RealIPExtractor{
		allowedHeaders: allowed,
		customHeaders:  custom,
	}
}

func NewDefaultRealIPExtractor() *RealIPExtractor {
	return NewRealIPExtractor(defaultHeaders...)
}

func (re *RealIPExtractor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if len(re.allowedHeaders) == 0 {
			next.ServeHTTP(rw, r)
			return
		}
		next.ServeHTTP(rw, r.WithContext(transport.WithRealIP(r.Context(), re.extractRemoteIP(r))))
	})
}

//goland:noinspection DuplicatedCode
func (re *RealIPExtractor) extractRemoteIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 1. Приоритет I: Эксклюзивные заголовки защищенных CDN/Провайдеров облаков
	if _, ok := re.allowedHeaders[HeaderCFConnectingIP]; ok {
		if ip := r.Header.Get(HeaderCFConnectingIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderTrueClientIP]; ok {
		if ip := r.Header.Get(HeaderTrueClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXCloudDeploymentUserIP]; ok {
		if ip := r.Header.Get(HeaderXCloudDeploymentUserIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderFastlyClientIP]; ok {
		if ip := r.Header.Get(HeaderFastlyClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXAzureClientIP]; ok {
		if ip := r.Header.Get(HeaderXAzureClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 2. Приоритет II: Стандартный X-Forwarded-For
	if _, ok := re.allowedHeaders[HeaderXForwardedFor]; ok {
		if xff := r.Header.Get(HeaderXForwardedFor); xff != "" {
			// strings.Cut разделяет строку по первому вхождению запятой без аллокаций кучи
			firstIP, _, _ := strings.Cut(xff, ",")
			if cleanIP := re.cleanAndValidateIP(firstIP); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 3. Приоритет III: Очищенный одиночный X-Real-IP
	if _, ok := re.allowedHeaders[HeaderXRealIP]; ok {
		if xrip := r.Header.Get(HeaderXRealIP); xrip != "" {
			if cleanIP := re.cleanAndValidateIP(xrip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 4. Приоритет IV: Альтернативные enterprise-заголовки прокси-систем
	if _, ok := re.allowedHeaders[HeaderClientIP]; ok {
		if ip := r.Header.Get(HeaderClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXClientIP]; ok {
		if ip := r.Header.Get(HeaderXClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}
	if _, ok := re.allowedHeaders[HeaderXClusterClientIP]; ok {
		if ip := r.Header.Get(HeaderXClusterClientIP); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 5. Приоритет V: Кастомная специфика (проверяем только то, что не вошло в списки выше)
	for _, header := range re.customHeaders {
		if ip := r.Header.Get(header); ip != "" {
			if cleanIP := re.cleanAndValidateIP(ip); cleanIP != "" {
				return cleanIP
			}
		}
	}

	// 6. Финальный фолбек: физический IP-адрес сокета
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if cleanIP := re.cleanAndValidateIP(host); cleanIP != "" {
			return cleanIP
		}
	}

	return strings.TrimSpace(r.RemoteAddr)
}

func (re *RealIPExtractor) cleanAndValidateIP(ip string) string {
	cleaned := strings.TrimSpace(ip)
	if cleaned == "" {
		return ""
	}
	if net.ParseIP(cleaned) != nil {
		return cleaned
	}
	return ""
}
