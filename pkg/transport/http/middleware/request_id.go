package middleware

import (
	"fmt"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

const (
	HeaderXRequestID     string = "X-Request-ID"
	HeaderXCorrelationID string = "X-Correlation-ID"
	HeaderRequestID      string = "Request-ID"
)

type RequestIDExtractor struct {
	headers []string
}

func NewRequestIDExtractor(headers []string) *RequestIDExtractor {
	return &RequestIDExtractor{
		headers: headers,
	}
}

func (re *RequestIDExtractor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(re.headers) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		var requestID string
		for _, header := range re.headers {
			requestID = r.Header.Get(header)
			if requestID != "" {
				break
			}
		}
		if requestID == "" {
			requestID = fmt.Sprintf("%s-%07d", transport.GetPrefix(), transport.NextReqID())
		}

		next.ServeHTTP(w, r.WithContext(transport.WithRequestID(r.Context(), requestID)))
	})
}
