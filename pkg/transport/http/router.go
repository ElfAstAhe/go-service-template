package http

import (
	"net/http"
)

type Router interface {
	GetRouter() http.Handler
}
