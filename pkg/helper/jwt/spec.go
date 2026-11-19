package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// TokenFactory определяет контракт для последовательной сборки,
// инициализации и криптографической подписи JWT-токенов.
type TokenFactory interface {
	// Claims выполняет валидацию бизнес-инвариантов и возвращает укомплектованный объект SubjectClaims.
	Claims(builder *SubjectClaimsBuilder) (*SubjectClaims, error)
	// Token собирает неподписанный объект jwt.Token на основе подготовленных утверждений Claims.
	Token(claims *SubjectClaims) (*jwt.Token, error)
	// TokenString подписывает токен и возвращает финальную компактную строку JWT.
	TokenString(token *jwt.Token) (string, error)
}

// TokenExtractor определяет контракт для извлечения, математической верификации
// и десериализации полезной нагрузки прилетающих JWT-токенов.
type TokenExtractor interface {
	// Claims извлекает и типизирует доменные утверждения SubjectClaims из верифицированного токена.
	Claims(token *jwt.Token) (*SubjectClaims, error)
	// Token осуществляет парсинг сырой текстовой строки токена и проверяет цифровую подпись.
	Token(tokenStr string) (*jwt.Token, error)
}
