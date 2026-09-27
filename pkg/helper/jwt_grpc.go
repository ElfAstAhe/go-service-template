package helper

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

// JWTGRPCHelper реализует специализированный транспортный адаптер (gRPC Bridge) для JWT-менеджера.
//
// Инкапсулирует рутину низкоуровневого извлечения строк авторизации из входящих HTTP/2 метаданных gRPC,
// позволяя gRPC-интерцепторам (Unary/Stream Interceptors) прозрачно валидировать сессии пользователей.
type JWTGRPCHelper struct {
	jwtHelper *JWTHelper // Ссылка на базовый криптографический JWT-менеджер фреймворка
}

// NewJWTGRPCHelper — фабричный конструктор gRPC-адаптера авторизации.
func NewJWTGRPCHelper(jwtHelper *JWTHelper) *JWTGRPCHelper {
	return &JWTGRPCHelper{
		jwtHelper: jwtHelper,
	}
}

// ExtractTokenStringFromMetadata вытаскивает сырую строку токена из переданной gRPC-карты метаданных (metadata.MD).
// Автоматически нормализует заголовок, срезая RFC-префикс «Bearer », если он присутствует.
func (jgh *JWTGRPCHelper) ExtractTokenStringFromMetadata(md metadata.MD, metadataName string) (string, error) {
	if strings.TrimSpace(metadataName) == "" {
		return "", errs.NewInvalidArgumentError("metadataName", "empty metadata name")
	}
	if md == nil {
		return "", errs.NewInvalidArgumentError("md", "nil metadata")
	}

	// Извлекаем срез значений по текстовому ключу (gRPC metadata ключи всегда в нижнем регистре)
	values := md.Get(metadataName)
	if len(values) == 0 {
		return "", nil
	}

	// Если префикс Bearer отсутствует — отдаем токен в чистом виде (Fallback совместимость)
	if !strings.HasPrefix(values[0], TokenPrefix) {
		return values[0], nil
	}

	// Мягко отсекаем префикс Bearer для передачи строки в парсер
	return strings.TrimPrefix(values[0], TokenPrefix), nil
}

// ExtractTokenStringFromContext извлекает входящие метаданные (Incoming Metadata) напрямую из context.Context
// gRPC-вызова и осуществляет безопасный поиск строки токена авторизации.
func (jgh *JWTGRPCHelper) ExtractTokenStringFromContext(ctx context.Context, metadataName string) (string, error) {
	if strings.TrimSpace(metadataName) == "" {
		return "", errs.NewInvalidArgumentError("metadataName", "empty metadata name")
	}

	// Безопасно извлекаем входящий транспортный gRPC-конверт
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errs.NewInvalidArgumentError("metadata", "no metadata found in incoming context")
	}

	res, err := jgh.ExtractTokenStringFromMetadata(md, metadataName)
	if err != nil {
		return "", errs.NewUtlJWTError("extract token string from metadata", err)
	}

	return res, nil
}

// ExtractTokenFromContext выполняет полный цикл gRPC-авторизации: извлекает строку токена из контекста метаданных
// и передает её в базовый JWT-хелпер для математической верификации цифровой подписи и сроков экспирации.
func (jgh *JWTGRPCHelper) ExtractTokenFromContext(ctx context.Context, metadataName string) (*jwt.Token, error) {
	tokenString, err := jgh.ExtractTokenStringFromContext(ctx, metadataName)
	if err != nil {
		return nil, errs.NewUtlJWTError("extract token string from context", err)
	}

	// Передаем строку в сквозной криптографический парсер
	return jgh.jwtHelper.ExtractTokenFromString(tokenString)
}
