package rest

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/hellofresh/health-go/v5"
)

type RouterOption func(*RouterOptions)

type RouterOptions struct {
	Address            string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	PrivateKeyPath     string
	CertificatePath    string
	Secure             bool
	MaxRequestBodySize int
	// telemetry options
	telemetryServiceName string
	// liveness and readiness
	livez  http.LivezFunc
	readyz http.ReadyzFunc
	// status handler
	statusHandler *health.Health
	//
	log logger.Logger
}

func NewRouterOptions() *RouterOptions {
	return &RouterOptions{}
}
