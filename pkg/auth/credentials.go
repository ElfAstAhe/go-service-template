package auth

import (
	"crypto/x509"

	"github.com/golang-jwt/jwt/v5"
)

// BasicCredentials инкапсулирует пару логин/пароль, извлечённую из HTTP Basic или gRPC metadata.
type BasicCredentials struct {
	Username string
	Password string
}

// CertCredentials инкапсулирует проверенный клиентский X.509-сертификат из TLS-соединения.
type CertCredentials struct {
	Certificate *x509.Certificate
}

// JWTCredentials инкапсулирует строку токена
type JWTCredentials struct {
	Token *jwt.Token
}
