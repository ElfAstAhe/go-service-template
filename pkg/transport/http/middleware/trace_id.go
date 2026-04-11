package middleware

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

const (
	HeaderXCloudTraceContext string = "X-Cloud-Trace-Context"
	HeaderTraceParent        string = "Traceparent"
	HeaderXTraceID           string = "X-Trace-ID"
	HeaderTraceID            string = "Trace-ID"
)

type TraceIDExtractor struct {
	headers []string
}

func NewTraceIDExtractor(headers []string) *TraceIDExtractor {
	return &TraceIDExtractor{
		headers: headers,
	}
}

func (te *TraceIDExtractor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(te.headers) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		var traceID string
		for _, header := range te.headers {
			traceID = r.Header.Get(header)
			if traceID != "" {
				break
			}
		}

		next.ServeHTTP(w, r.WithContext(transport.WithTraceID(r.Context(), traceID)))
	})
}
