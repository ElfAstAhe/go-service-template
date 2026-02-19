package transport

import (
	"net/http"
)

type HTTPRouter interface {
	GetRouter() http.Handler
}
