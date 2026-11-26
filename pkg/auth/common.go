package auth

// Рантайм-константы имен заголовков, кук и gRPC-метаданных по умолчанию.
const (
	// DefaultHeaderName auth header param name
	DefaultHeaderName string = "Authorization"

	// DefaultCookieName auth cookie name
	DefaultCookieName string = "Authorization"

	// DefaultMetadataName auth meta data name
	DefaultMetadataName string = "authorization"
)

const (
	// DefaultBasicPrefix auth basic prefix
	DefaultBasicPrefix string = "Basic "

	// DefaultJWTPrefix auth jwt prefix
	DefaultJWTPrefix string = "Bearer "
)
