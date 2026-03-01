package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade"
	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	mware "github.com/ElfAstAhe/go-service-template/pkg/transport/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hellofresh/health-go/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swagh "github.com/swaggo/http-swagger"
)

type AppChiRouter struct {
	router     *chi.Mux
	log        logger.Logger
	config     *conf.HTTPConfig
	health     *health.Health
	healthz    transport.HealthzFunc
	readyz     transport.ReadyzFunc
	testFacade facade.TestFacade
}

func NewAppChiRouter(
	config *conf.HTTPConfig,
	logger logger.Logger,
	health *health.Health,
	healthz transport.HealthzFunc,
	readyz transport.ReadyzFunc,
	testFacade facade.TestFacade,
) *AppChiRouter {
	res := &AppChiRouter{
		router:     chi.NewRouter(),
		log:        logger,
		config:     config,
		health:     health,
		healthz:    healthz,
		readyz:     readyz,
		testFacade: testFacade,
	}

	// setup middleware
	res.setupMiddleware(logger)

	// mount debug
	res.router.Mount("/debug", middleware.Profiler())
	// mount swagger
	res.router.Mount("/swagger/", swagh.WrapHandler)
	// mount status
	res.router.Mount("/status", res.health.Handler())
	// mount metrics
	res.router.Mount("/metrics", promhttp.Handler())

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
	// compress (add any content-types)
	cr.router.Use(mware.NewHTTPCompress(logger,
		"application/json", "plain/text",
	).Handle)
	// decompress
	cr.router.Use(mware.NewHTTPDecompress(int64(cr.config.MaxRequestBodySize), logger).Handle)
	// income/outcome logger
	cr.router.Use(mware.NewHTTPRequestLogger(logger).Handle)
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
		r.Route("/test", func(r chi.Router) {
			r.Get("/{id}", cr.getAPITest)
			r.Get("/search", cr.getAPITestSearch)
			r.Get("/", cr.getAPITestList)
			r.Post("/", cr.postAPITest)
			r.Put("/{id}", cr.putAPITest)
			r.Delete("/{id}", cr.deleteAPITest)
		})
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
