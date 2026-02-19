package middleware

import (
	"io"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/andybalholm/brotli"
	"github.com/go-chi/chi/v5/middleware"
)

// DefaultCompressionLevel Default compression level
const DefaultCompressionLevel = 5

// Accept-Encoding, Content-Encoding
const (
	encodingBrotli = "br" // brotli
)

type HTTPCompress struct {
	compressor          *middleware.Compressor
	allowedContentTypes []string
	log                 logger.Logger
}

func NewHTTPCompress(logger logger.Logger, allowedContentTypes ...string) *HTTPCompress {
	// create instance
	res := &HTTPCompress{
		allowedContentTypes: allowedContentTypes,
		log:                 logger.GetLogger("http_compress_middleware"),
	}
	// init instance
	res.init()

	return res
}

func (hc *HTTPCompress) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hc.log.Debug("HTTPCompress.Handle start")
		defer hc.log.Debug("HTTPCompress.Handle finish")

		hc.compressor.Handler(next).ServeHTTP(w, r)
	})
}

func (hc *HTTPCompress) init() {
	hc.compressor = middleware.NewCompressor(DefaultCompressionLevel, hc.allowedContentTypes...)
	hc.compressor.SetEncoder(encodingBrotli, hc.brotliWriterFactory)
}

func (hc *HTTPCompress) brotliWriterFactory(w io.Writer, level int) io.Writer {
	return brotli.NewWriterLevel(w, level)
}
