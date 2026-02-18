package rest

import (
	_ "expvar"
	"net/http"

	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hellofresh/health-go/v5"
	swagh "github.com/swaggo/http-swagger"
)

type AppChiRouter struct {
	router  *chi.Mux
	log     logger.Logger
	config  *conf.HTTPConfig
	health  *health.Health
	healthz transport.HealthzFunc
	readyz  transport.ReadyzFunc
}

func NewAppChiRouter(
	config *conf.HTTPConfig,
	logger logger.Logger,
	health *health.Health,
	healthz transport.HealthzFunc,
	readyz transport.ReadyzFunc,
) *AppChiRouter {
	res := &AppChiRouter{
		router:  chi.NewRouter(),
		log:     logger,
		config:  config,
		health:  health,
		healthz: healthz,
		readyz:  readyz,
	}

	// setup middleware
	res.setupMiddleware(logger)

	// mount debug
	res.router.Mount("/debug", middleware.Profiler())
	// mount swagger
	res.router.Mount("/swagger/", swagh.WrapHandler)
	// mount status
	res.router.Mount("/status", res.health.Handler())

	// setup routes
	res.setupRoutes()

	return res
}

func (cr *AppChiRouter) GetRouter() http.Handler {
	return cr.router
}

func (cr *AppChiRouter) setupMiddleware(logger logger.Logger) {
	// jwt auth extractor - extract user info from token
	// .. cr.router.Use(appmware.NewAuthExtractorMiddleware(cr.authHelper, cr.jwtHTTPHelper, logger).Handle)
	// requestID
	cr.router.Use(middleware.RequestID)
	// realIP
	cr.router.Use(middleware.RealIP)
	// compress
	// ..
	// decompress
	// ..
	// income/outcome logger
	// ..
	// recoverer
	cr.router.Use(middleware.Recoverer)
	// timeout
	cr.router.Use(middleware.Timeout(cr.config.ReadTimeout))
}

func (cr *AppChiRouter) setupRoutes() {
	// health check
	cr.router.Get("/healthz", cr.getHealthz)
	// readiness check
	cr.router.Get("/readyz", cr.getReadyz)

	// api
	cr.router.Route("/api", func(r chi.Router) {
		/*
			// auth sub-router
			r.Route("/auth", func(r chi.Router) {
				r.Post("/login", cr.postAPIAuthLogin)       // POST /api/auth/login
				r.Post("/register", cr.postAPIAuthRegister) // POST /api/auth/register
			})
			// users sub-router
			r.Route("/users", func(r chi.Router) {
				r.Get("/profile", cr.getAPIUsersProfile)
				r.Put("/keys", cr.putAPIUsersKeys)
				r.Put("/password", cr.putAPIUsersPassword)
			})

		*/
	})
}
