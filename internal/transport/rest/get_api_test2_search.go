package rest

import (
	"net/http"

	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITest2Search обработчик пути GET /api/test2/search?code=...
//
// @Summary      Получить
// @Description  Удаляет запись по её ID (Soft Delete)
// @Tags         test
// @Produce      json
// @Param        code   query      string  true  "code записи" format(string)
// @Success      200  {object}  Test2 "Тестовые данные 2"
// @Failure      404  {object}  ErrorDTO "Запись не найдена"
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test2/search [get]
func (cr *AppChiRouter) getAPITest2Search(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("getAPITest2Search start, requestID [%s]", pkgtr.RequestID(r.Context()))
	defer cr.log.Debugf("getAPITest2Search finish, requestID [%s]", pkgtr.RequestID(r.Context()))

	code, err := pkghttp.GetQueryString(r, "code")
	if err != nil {
		pkghttp.RenderError(rw, err, pkghttp.MapToHTTPStatus)

		return
	}

	res, err := cr.test2Facade.GetByCode(r.Context(), code)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
