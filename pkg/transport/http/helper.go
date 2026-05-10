package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

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

func GetQueryIntDefault(r *http.Request, key string, defaultValue int) int {
	res, err := GetQueryInt(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

func GetQueryString(r *http.Request, key string) (string, error) {
	res := r.URL.Query().Get(key)
	if res == "" {
		return "", errs.NewInvalidArgumentError(key, "empty or not exists")
	}

	return res, nil
}

func GetQueryStringDefault(r *http.Request, key string, defaultValue string) string {
	res, err := GetQueryString(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

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

func GetQueryBoolDefault(r *http.Request, key string, defaultValue bool) bool {
	res, err := GetQueryBool(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

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

func GetQueryTimeDefault(r *http.Request, key string, defaultValue time.Time) time.Time {
	res, err := GetQueryTime(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

func GetQueryStringArray(r *http.Request, key string) ([]string, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil, errs.NewInvalidArgumentError(key, "empty or not exists")
	}

	return strings.Split(val, ","), nil
}

func GetQueryStringArrayDefault(r *http.Request, key string, defaultValue []string) []string {
	res, err := GetQueryStringArray(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

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

func GetQueryIntArrayDefault(r *http.Request, key string, defaultValue []int) []int {
	res, err := GetQueryIntArray(r, key)
	if err != nil {
		return defaultValue
	}

	return res
}

func DecodeJSON(r *http.Request, dst any) error {
	// 1. Ограничиваем чтение (например, 1Мб), чтобы не выесть RAM
	// MaxBytesReader вернет ошибку, если тело больше лимита
	r.Body = http.MaxBytesReader(nil, r.Body, 1024*1024)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	// 2. Strict mode: если клиент прислал поле, которого нет в DTO — это 400
	// Помогает отловить опечатки на фронте (например, "iddd" вместо "id")
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return errs.NewInvalidArgumentErrorChain("body", "invalid_json_format", err)
	}

	return nil
}
