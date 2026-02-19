package middleware

import (
	"net/http"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTPRequestLogger struct {
	log logger.Logger
}

func NewHTTPRequestLogger(logger logger.Logger) *HTTPRequestLogger {
	return &HTTPRequestLogger{
		log: logger.GetLogger("http_request_logger"),
	}
}

func (hrl *HTTPRequestLogger) Handle(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hrl.log.Debug("HTTPRequestLogger.Handle start")
		defer hrl.log.Debug("HTTPRequestLogger.Handle end")

		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			hrl.log.DebugW("http request",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"uri", r.RequestURI,
				"status", wrw.Status(),
				"latency", time.Since(start),
				"bytes", wrw.BytesWritten(),
				"remote_ip", r.RemoteAddr,
			)
		}()

		handler.ServeHTTP(wrw, r)
	})
}
