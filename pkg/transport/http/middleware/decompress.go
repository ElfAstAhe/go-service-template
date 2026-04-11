package middleware

import (
	"net/http"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type Decompress struct {
	maxRequestBodySize int64
	log                logger.Logger
}

func NewDecompress(maxRequestBodySize int64, logger logger.Logger) *Decompress {
	return &Decompress{
		maxRequestBodySize: maxRequestBodySize,
		log:                logger.GetLogger("http_decompress_middleware"),
	}
}

func (hd *Decompress) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		hd.log.Debug("HTTPDecompress start")
		defer hd.log.Debug("HTTPDecompress finish")

		encoding := strings.ToLower(r.Header.Get("Content-Encoding"))
		if encoding != "" {
			hd.log.Debugf("HTTPDecompress encoding: [%s]", encoding)

			if hd.maxRequestBodySize > 0 {
				hd.log.Debugf("HTTPDecompress maxRequestBodySize [%d] setting applied", hd.maxRequestBodySize)
				r.Body = http.MaxBytesReader(rw, r.Body, hd.maxRequestBodySize)
			}

			dr, err := utils.NewDecompressReader(encoding, r.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)

				return
			}

			r.Body = dr
		}

		next.ServeHTTP(rw, r)
	})
}
