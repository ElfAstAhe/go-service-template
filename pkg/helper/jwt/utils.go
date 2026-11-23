package jwt

import (
	"fmt"
	"slices"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

// GetSigningMethod возвращает объект jwt.SigningMethod на основе его текстового имени (алгоритма).
// Если метод неизвестен библиотеке, возвращает кастомную ошибку *errs.UtlJWTError.
func GetSigningMethod(alg string) (jwt.SigningMethod, error) {
	res := jwt.GetSigningMethod(alg)
	if utils.IsNil(res) {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("unknown signing method [%s]", alg), nil)
	}

	return res, nil
}

// IsSigningMethodExists проверяет, зарегистрирован ли указанный алгоритм подписи в рантайме библиотеки jwt.
func IsSigningMethodExists(alg string) bool {
	return slices.Contains(jwt.GetAlgorithms(), alg)
}
