package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func (cr *AppChiRouter) renderError(rw http.ResponseWriter, err error) {
	status := mapToHTTPStatus(err)

	if status >= http.StatusInternalServerError {
		cr.renderEmpty(rw, status)
	} else {
		cr.renderJSON(rw, status, transport.NewErrorDTOFromError(status, err))
	}
}

func (cr *AppChiRouter) renderJSON(rw http.ResponseWriter, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	_, _ = rw.Write(js)
}

func (cr *AppChiRouter) renderEmpty(rw http.ResponseWriter, status int) {
	rw.WriteHeader(status)
}

func (cr *AppChiRouter) getQueryInt(r *http.Request, key string, defaultValue int) (int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue, nil
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, errs.NewInvalidArgumentErrorChain(key, val, err)
	}

	return intVal, nil
}

func (cr *AppChiRouter) getQueryString(r *http.Request, key string, defaultValue string) string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}

	return val
}

func (cr *AppChiRouter) getQueryBool(r *http.Request, key string, defaultValue bool) (bool, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue, nil
	}
	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return false, errs.NewInvalidArgumentErrorChain(key, val, err)
	}

	return boolVal, nil
}

func (cr *AppChiRouter) getQueryTime(r *http.Request, key string, defaultValue time.Time) (time.Time, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue, nil
	}
	timeVal, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, errs.NewInvalidArgumentErrorChain(key, val, err)
	}

	return timeVal, nil
}

func (cr *AppChiRouter) getQueryStringArray(r *http.Request, key string, defaultValue []string) []string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}

	return strings.Split(val, ",")
}

func (cr *AppChiRouter) getQueryIntArray(r *http.Request, key string, defaultValue []int) ([]int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue, nil
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

func (cr *AppChiRouter) decodeJSON(r *http.Request, dst any) error {
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
