package http

// HealthzFunc health check
type HealthzFunc func() bool

// ReadyzFunc readiness check
type ReadyzFunc func() bool

// MapToHTTPStatusFunc error mapper to http status code
type MapToHTTPStatusFunc func(error) int
