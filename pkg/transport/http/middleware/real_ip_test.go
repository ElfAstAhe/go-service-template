package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тестовый хендлер для перехвата IP из контекста
func newTestHandler(capturedIP *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем получение IP из вашего пакета transport
		// Для автономности теста проверяем значение напрямую через ключ или кастомную логику,
		// но здесь мы просто вызовем extractRemoteIP напрямую или симулируем контекст.
		w.WriteHeader(http.StatusOK)
	})
}

func TestRealIPExtractor_ExtractRemoteIP(t *testing.T) {
	tests := []struct {
		name           string
		allowedHeaders []string
		remoteAddr     string
		headers        map[string]string
		wantIP         string
	}{
		{
			name:           "Дефолтный экстрактор: берет первый IP из X-Forwarded-For",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178",
			},
			wantIP: "203.0.113.195",
		},
		{
			name:           "Дефолтный экстрактор: X-Forwarded-For с пробелами",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "  203.0.113.195  , 70.41.3.18",
			},
			wantIP: "203.0.113.195",
		},
		{
			name:           "Дефолтный экстрактор: фолбек на X-Real-IP, если XFF пустой",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.1",
			},
			wantIP: "198.51.100.1",
		},
		{
			name:           "Защита от спуфинга: игнорирует CF-Connecting-IP, если он не разрешен",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-Real-IP":        "198.51.100.1",
			},
			wantIP: "198.51.100.1",
		},
		{
			name:           "Явное включение CDN: приоритет Cloudflare над XFF",
			allowedHeaders: []string{HeaderCFConnectingIP, HeaderXForwardedFor},
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-Forwarded-For":  "203.0.113.195",
			},
			wantIP: "1.1.1.1",
		},
		{
			name:           "Валидация: игнорирует некорректный IP в заголовке и идет по каскаду ниже",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "not-an-ip, 70.41.3.18",
				"X-Real-IP":       "198.51.100.1",
			},
			wantIP: "198.51.100.1",
		},
		{
			name:           "Кастомный заголовок: корректная обработка",
			allowedHeaders: []string{"X-Custom-Private-IP"},
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"X-Custom-Private-IP": "10.0.0.5",
			},
			wantIP: "10.0.0.5",
		},
		{
			name:           "Фолбек на RemoteAddr: когда все заголовки отсутствуют",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.0.2.1:54321",
			headers:        map[string]string{},
			wantIP:         "192.0.2.1",
		},
		{
			name:           "Фолбек на RemoteAddr без порта (например, unix-сокет)",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.0.2.2",
			headers:        map[string]string{},
			wantIP:         "192.0.2.2",
		},
		{
			name:           "Поддержка IPv6 в X-Forwarded-For",
			allowedHeaders: defaultHeaders,
			remoteAddr:     "192.168.1.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1, 192.0.2.1",
			},
			wantIP: "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewRealIPExtractor(tt.allowedHeaders...)
			req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			req.RemoteAddr = tt.remoteAddr

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			gotIP := extractor.extractRemoteIP(req)
			if gotIP != tt.wantIP {
				t.Errorf("extractRemoteIP() = %q, want %q", gotIP, tt.wantIP)
			}
		})
	}
}

func TestRealIPExtractor_Handler_EmptyHeaders(t *testing.T) {
	// Если передать пустой список заголовков, мидлварь должен просто пропустить запрос дальше
	extractor := NewRealIPExtractor()
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	rr := httptest.NewRecorder()

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	extractor.Handler(nextHandler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("Следующий хендлер в цепочке не был вызван")
	}
}
