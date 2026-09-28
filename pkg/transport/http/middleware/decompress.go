package middleware

import (
	"net/http"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Decompress реализует промежуточный обработчик (Middleware) автоматической распаковки
// входящего сжатого REST/HTTP трафика (Gzip / Brotli) со стороны клиентов.
//
// Инкапсулирует политики защиты памяти от Zip-бомб посредством жесткого лимитирования
// размера входящего потока и динамически подменяет r.Body на декодирующий ридер.
type Decompress struct {
	maxRequestBodySize int64         // Жесткий верхний лимит (в байтах) на максимальный размер тела запроса (Защита от DoS/OOM)
	log                logger.Logger // Изолированный структурированный логгер компонента
}

// NewDecompress — фабричный конструктор и инициализатор middleware декомпрессии трафика.
func NewDecompress(maxRequestBodySize int64, logger logger.Logger) *Decompress {
	return &Decompress{
		maxRequestBodySize: maxRequestBodySize,
		log:                logger.GetLogger("http_decompress_middleware"),
	}
}

// Handle осуществляет перехват входящего HTTP-запроса (http.Handler Middleware Pattern).
// Анализирует заголовок Content-Encoding, проактивно защищает RAM через MaxBytesReader,
// инжектирует распаковывающий поток и передает управление дальше по цепочке обработчиков.
func (hd *Decompress) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		hd.log.Debug("HTTPDecompress start")
		defer hd.log.Debug("HTTPDecompress finish")

		encoding := strings.ToLower(r.Header.Get("Content-Encoding"))
		if encoding != "" {
			hd.log.Debugf("HTTPDecompress encoding: [%s]", encoding)

			// 🛡️ Защитный барьер: лимитируем чтение байт ДО начала распаковки для предотвращения Zip Bomb атак
			if hd.maxRequestBodySize > 0 {
				hd.log.Debugf("HTTPDecompress maxRequestBodySize [%d] setting applied", hd.maxRequestBodySize)
				r.Body = http.MaxBytesReader(rw, r.Body, hd.maxRequestBodySize)
			}

			// Создаем специализированный декомпрессирующий поток через утилитарный хелпер фреймворка
			dr, err := utils.NewDecompressReader(encoding, r.Body)
			if err != nil {
				// В случае отправки неподдерживаемого или битого сжатия возвращаем 400 Bad Request
				http.Error(rw, err.Error(), http.StatusBadRequest)

				return
			}

			// Декорируем исходное тело запроса распаковывающим ридером
			r.Body = dr
		}

		next.ServeHTTP(rw, r)
	})
}
