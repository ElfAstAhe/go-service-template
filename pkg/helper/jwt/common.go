package jwt

import (
	"time"
)

// Рантайм-дефолты для безопасной конфигурации JWT-токенов в экосистеме.
const (
	DefaultSigningMethodName  string        = "HS256"          // Метод подписи по умолчанию: HMAC-SHA256
	DefaultExpirationInterval time.Duration = 15 * time.Minute // Стандартное время жизни Access-токена (15 минут)
)
