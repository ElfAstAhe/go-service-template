package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// putAPITest обработчик пути PUT /api/test/{id}
//
// @Summary      Изменение тестовых данных
// @Description  Изменяет тестовые данные
// @Tags         test
// @Accept       json
// @Produce      json
// @Param        id     path      string   true  "ID записи" format(string)
// @Param        input  body      Test  true  "Тестовые данные"
// @Success      200    {object}  Test
// @Failure      400    {object}  ErrorDTO
// @Failure      404    {object}  ErrorDTO
// @Failure      409    {object}  ErrorDTO
// @Failure      500    "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test/{id} [put]
//
//goland:noinspection DuplicatedCode
func (cr *AppChiRouter) putAPITest(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("putAPITest start, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)
	defer cr.log.Debugf("putAPITest finish, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)

	var income = &dto.Test{}
	err := pkghttp.DecodeJSON(r, income)
	if err != nil {
		pkghttp.RenderError(rw, err, pkghttp.MapToHTTPStatus)

		return
	}

	res, err := cr.testFacade.Change(r.Context(), id, income)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
