package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// GetQueryInt извлекает параметр строки запроса по ключу и преобразует его в тип int.
// Возвращает InvalidArgumentError, если параметр отсутствует или не может быть распарсен как число.
func GetQueryInt(r *http.Request, key string) (int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return 0, errs.NewInvalidArgumentError(key, "empty or not exists")
	}
	res, err := strconv.Atoi(val)
	if err != nil {
		return 0, errs.NewInvalidArgumentErrorChain(key, val, err)
	}

	return res, nil
}

// GetQueryIntDefault извлекает целочисленный параметр строки запроса, возвращая defaultValue в случае его отсутствия или ошибки парсинга.
func GetQueryIntDefault(r *http.Request, key string, defaultValue int) int {
	res, err := GetQueryInt(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

// GetQueryString извлекает строковый параметр из строки запроса.
// Возвращает InvalidArgumentError, если параметр пуст или отсутствует.
func GetQueryString(r *http.Request, key string) (string, error) {
	res := r.URL.Query().Get(key)
	if res == "" {
		return "", errs.NewInvalidArgumentError(key, "empty or not exists")
	}

	return res, nil
}

// GetQueryStringDefault извлекает строковый параметр строки запроса, возвращая defaultValue в случае его отсутствия.
func GetQueryStringDefault(r *http.Request, key string, defaultValue string) string {
	res, err := GetQueryString(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

// GetQueryBool извлекает параметр строки запроса и интерпретирует его как логический тип bool (true/false).
// Поддерживает стандартные литералы strconv.ParseBool. Возвращает ошибку при невалидном синтаксисе.
func GetQueryBool(r *http.Request, key string) (bool, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return false, errs.NewInvalidArgumentError(key, "empty or not exists")
	}
	res, err := strconv.ParseBool(val)
	if err != nil {
		return false, errs.NewInvalidArgumentErrorChain(key, val, err)
	}

	return res, nil
}

// GetQueryBoolDefault извлекает логический параметр строки запроса, возвращая defaultValue в случае его отсутствия или сбоя.
func GetQueryBoolDefault(r *http.Request, key string, defaultValue bool) bool {
	res, err := GetQueryBool(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

// GetQueryTime извлекает временной параметр строки запроса и парсит его согласно строгому стандарту ISO-8601 (time.RFC3339).
// Если параметр отсутствует, возвращает системный маркер utils.ZeroTime без генерации ошибки.
func GetQueryTime(r *http.Request, key string) (time.Time, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return utils.ZeroTime, nil
	}
	res, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, errs.NewInvalidArgumentErrorChain(key, val, err)
	}

	return res, nil
}

// GetQueryTimeDefault извлекает временной параметр, возвращая defaultValue в случае синтаксической ошибки парсинга таймштампа.
func GetQueryTimeDefault(r *http.Request, key string, defaultValue time.Time) time.Time {
	res, err := GetQueryTime(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

// GetQueryStringArray извлекает параметр строки запроса, расщепляя его через запятую на плоский строковый срез []string.
func GetQueryStringArray(r *http.Request, key string) ([]string, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil, errs.NewInvalidArgumentError(key, "empty or not exists")
	}

	return strings.Split(val, ","), nil
}

// GetQueryStringArrayDefault извлекает массив строк, возвращая defaultValue в случае отсутствия оригинального параметра.
func GetQueryStringArrayDefault(r *http.Request, key string, defaultValue []string) []string {
	res, err := GetQueryStringArray(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

// GetQueryIntArray извлекает переданный через запятую список значений и преобразует их в целочисленный срез []int.
// Оптимизировано: сразу выполняет предварительное выделение памяти (pre-allocation) под итоговую коллекцию.
func GetQueryIntArray(r *http.Request, key string) ([]int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil, errs.NewInvalidArgumentError(key, "empty or not exists")
	}
	strArr := strings.Split(val, ",")
	intArr := make([]int, len(strArr))
	var err error
	for i, str := range strArr {
		intArr[i], err = strconv.Atoi(str)
		if err != nil {
			return nil, errs.NewInvalidArgumentErrorChain(key, val, err)
		}
	}

	return intArr, nil
}

// GetQueryIntArrayDefault извлекает массив чисел, возвращая дефолтный срез defaultValue при провале валидации или парсинга.
func GetQueryIntArrayDefault(r *http.Request, key string, defaultValue []int) []int {
	res, err := GetQueryIntArray(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}
