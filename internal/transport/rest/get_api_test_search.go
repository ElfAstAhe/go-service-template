package rest

import (
	"net/http"

	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITest godoc
// @Summary      Получить
// @Description  Удаляет запись по её ID (Soft Delete)
// @Tags         test
// @Produce      json
// @Param        code   query      string  true  "code записи" format(string)
// @Success      200  {object}  TestDTO "Тестовые данные"
// @Failure      404  {object}  ErrorDTO "Запись не найдена"
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test/search [get]
func (cr *AppChiRouter) getAPITestSearch(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("getAPITestSearch start, requestID [%s]", middleware.GetReqID(r.Context()))
	defer cr.log.Debugf("getAPITestSearch finish, requestID [%s]", middleware.GetReqID(r.Context()))

	code, err := pkghttp.GetQueryString(r, "code")
	if err != nil {
		pkghttp.RenderError(rw, err, pkghttp.MapToHTTPStatus)

		return
	}

	res, err := cr.testFacade.GetByCode(r.Context(), code)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
