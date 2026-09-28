package middleware

import (
	"io"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/andybalholm/brotli"
	"github.com/go-chi/chi/v5/middleware"
)

// DefaultCompressionLevel определяет дефолтную степень сжатия данных алгоритмами компрессии (баланс скорость/плотность).
const DefaultCompressionLevel = 5

// Константные строковые литералы HTTP-спецификаций кодирования контента.
const (
	encodingBrotli = "br" // Идентификатор алгоритма Brotli в HTTP-заголовках Accept-Encoding / Content-Encoding
)

// Compress реализует промежуточный обработчик (Middleware) динамического сжатия ответов HTTP-сервера.
//
// Инкапсулирует под капотом chi-компрессор, расширяя его нативной поддержкой высокопроизводительного
// алгоритма Brotli (br). Позволяет кратно сократить объем сетевого REST/HTTP-трафика для текстовых JSON-ответов.
type Compress struct {
	compressor          *middleware.Compressor // Ссылка на низкоуровневый движок компрессии chi
	allowedContentTypes []string               // Вайтлист MIME-типов контента, разрешенных для сжатия
	log                 logger.Logger          // Изолированный структурированный логгер компонента
}

// NewCompress — фабричный конструктор и инициализатор middleware сжатия трафика.
func NewCompress(logger logger.Logger, allowedContentTypes ...string) *Compress {
	// create instance
	res := &Compress{
		allowedContentTypes: allowedContentTypes,
		log:                 logger.GetLogger("http_compress_middleware"),
	}
	// init instance
	res.init()

	return res
}

// Handle встраивает компрессор в каскадную цепочку обработки HTTP-запросов (http.Handler Middleware Pattern).
// Перехватывает поток записи ответа (http.ResponseWriter), динамически инжектируя туда Brotli/Gzip врайтер
// на основе входящих заголовков Accept-Encoding, переданных клиентом (браузером, мобильным приложением).
func (hc *Compress) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hc.log.Debug("HTTPCompress.Handle start")
		defer hc.log.Debug("HTTPCompress.Handle finish")

		// Делегируем низкоуровневое управление и парсинг заголовков встроенному компрессору
		hc.compressor.Handler(next).ServeHTTP(w, r)
	})
}

// init — служебный (неэкспортируемый) метод предварительной конфигурации и регистрации Brotli-фабрики.
func (hc *Compress) init() {
	hc.compressor = middleware.NewCompressor(DefaultCompressionLevel, hc.allowedContentTypes...)
	// Переопределяем стандартный реестр кодировщиков chi, добавляя поддержку Brotli
	hc.compressor.SetEncoder(encodingBrotli, hc.brotliWriterFactory)
}

// brotliWriterFactory — служебный (неэкспортируемый) фабричный метод генерации Brotli-потока.
func (hc *Compress) brotliWriterFactory(w io.Writer, level int) io.Writer {
	return brotli.NewWriterLevel(w, level)
}
