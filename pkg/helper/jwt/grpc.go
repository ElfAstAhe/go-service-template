package jwt

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

const (
	// DefaultGRPCTokenPrefix определяет стандартный префикс схемы авторизации в метаданных gRPC.
	DefaultGRPCTokenPrefix = "Bearer "
)

// GRPCExtractor инкапсулирует логику извлечения и первичного парсинга JWT-токенов
// из метаданных входящих gRPC-запросов.
type GRPCExtractor struct {
	tokenExtractor TokenExtractor
	tokenPrefix    string
}

// NewGRPCExtractor создает новый экземпляр GRPCExtractor с заданным экстрактором токенов и префиксом.
func NewGRPCExtractor(tokenExtractor TokenExtractor, tokenPrefix string) *GRPCExtractor {
	return &GRPCExtractor{
		tokenExtractor: tokenExtractor,
		tokenPrefix:    tokenPrefix,
	}
}

// TokenFromMeta извлекает JWT-токен из переданного объекта gRPC metadata.MD по имени ключа и проверяет его подпись.
func (ge *GRPCExtractor) TokenFromMeta(md metadata.MD, name string) (*jwt.Token, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.NewInvalidArgumentError("metadataName", "is required")
	}
	if md == nil {
		return nil, errs.NewInvalidArgumentError("md", "is required")
	}
	tokenStr, err := ge.tokenStrFromMeta(md, name)
	if err != nil {
		return nil, errs.NewUtlJWTError("extract token string from metadata failed", err)
	}
	res, err := ge.tokenExtractor.Token(tokenStr)
	if err != nil {
		return nil, errs.NewUtlJWTError("token extraction failed", err)
	}

	return res, nil
}

// TokenFromContext извлекает JWT-токен из входящего gRPC-контекста по имени ключа метаданных и проверяет его подпись.
func (ge *GRPCExtractor) TokenFromContext(ctx context.Context, name string) (*jwt.Token, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.NewInvalidArgumentError("name", "is required")
	}
	if utils.IsNil(ctx) {
		return nil, errs.NewInvalidArgumentError("context", "is required")
	}

	tokenStr, err := ge.tokenStrFromContext(ctx, name)
	if err != nil {
		return nil, errs.NewUtlJWTError("extract token string from context failed", err)
	}
	res, err := ge.tokenExtractor.Token(tokenStr)
	if err != nil {
		return nil, errs.NewUtlJWTError("token extraction failed", err)
	}

	return res, nil
}

func (ge *GRPCExtractor) tokenStrFromMeta(md metadata.MD, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errs.NewInvalidArgumentError("metadataName", "is required")
	}
	if md == nil {
		return "", errs.NewInvalidArgumentError("md", "is required")
	}

	// Извлекаем срез значений по текстовому ключу (gRPC metadata ключи всегда в нижнем регистре)
	values := md.Get(name)
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return "", errs.NewUtlJWTError("no token data", nil)
	}

	// Если префикс Bearer отсутствует — отдаем токен в чистом виде (Fallback совместимость)
	if !strings.HasPrefix(values[0], ge.tokenPrefix) {
		return values[0], nil
	}

	// Мягко отсекаем префикс Bearer для передачи строки в парсер
	return strings.TrimSpace(strings.TrimPrefix(values[0], ge.tokenPrefix)), nil
}

func (ge *GRPCExtractor) tokenStrFromContext(ctx context.Context, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errs.NewInvalidArgumentError("name", "is required")
	}
	if utils.IsNil(ctx) {
		return "", errs.NewInvalidArgumentError("context", "is required")
	}

	// Безопасно извлекаем входящий транспортный gRPC-конверт
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errs.NewUtlJWTError("no token source data", nil)
	}

	res, err := ge.tokenStrFromMeta(md, name)
	if err != nil {
		return "", errs.NewUtlJWTError("extract token string from metadata failed", err)
	}

	return res, nil
}
