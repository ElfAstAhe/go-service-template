package http

import (
	"encoding/json"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
)

// RenderError осуществляет сериализацию и отправку ошибок клиенту с автоматической маскировкой 5xx статусов.
// Если переданный errorMapper транслирует ошибку в код >= 500, метод выполняет безопасный пустой рендеринг,
// полностью исключая случайную утечку чувствительных системных данных и инфраструктурных трейсов на фронтенд.
func RenderError(rw http.ResponseWriter, err error, errorMapper MapToHTTPStatusFunc) {
	status := errorMapper(err)

	if status >= http.StatusInternalServerError {
		RenderEmpty(rw, status)
	} else {
		RenderJSON(rw, status, transport.NewErrorDTOFromError(status, err), errorMapper)
	}
}

// RenderErrorDefault выполняет сериализацию ошибки, используя системный маппер по умолчанию (MapToHTTPStatus).
func RenderErrorDefault(rw http.ResponseWriter, err error) {
	RenderError(rw, err, MapToHTTPStatus)
}

// RenderJSON выполняет маршалинг произвольных структур (DTO/Доменных моделей) в формат JSON
// и отправляет их клиенту с установкой заголовка Content-Type и соответствующего HTTP-статуса.
// В случае сбоя сериализации данных автоматически перехватывает управление и инициирует рендеринг ошибки.
func RenderJSON(rw http.ResponseWriter, status int, data any, errorMapper MapToHTTPStatusFunc) {
	js, err := json.Marshal(data)
	if err != nil {
		RenderError(rw, err, errorMapper)

		return
	}

	rw.Header().Set("Content-Type", MediaTypeApplicationJSON+";charset=utf-8")
	rw.WriteHeader(status)
	_, _ = rw.Write(js)
}

// RenderJSONDefault выполняет стандартную JSON-сериализацию данных с использованием маппера по умолчанию (MapToHTTPStatus).
func RenderJSONDefault(rw http.ResponseWriter, status int, data any) {
	RenderJSON(rw, status, data, MapToHTTPStatus)
}

// RenderEmpty выполняет быструю отправку пустого тела ответа, фиксируя в сетевом пакете только переданный HTTP-статус.
func RenderEmpty(rw http.ResponseWriter, status int) {
	rw.WriteHeader(status)
}
