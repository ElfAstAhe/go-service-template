package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// putAPITest godoc
// @Summary      Изменение тестовых данных
// @Description  Изменяет тестовые данные
// @Tags         test
// @Accept       json
// @Produce      json
// @Param        id     path      string   true  "ID записи" format(string)
// @Param        input  body      TestDTO  true  "Тестовые данные"
// @Success      200    {object}  TestDTO
// @Failure      400    {object}  ErrorDTO
// @Failure      404    {object}  ErrorDTO
// @Failure      409    {object}  ErrorDTO
// @Failure      500    "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test/{id} [put]
func (cr *AppChiRouter) putAPITest(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("putAPITest start, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)
	defer cr.log.Debugf("putAPITest finish, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)

	var income = &dto.TestDTO{}
	err := cr.decodeJSON(r, income)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	res, err := cr.testFacade.Change(r.Context(), id, income)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	cr.renderJSON(rw, http.StatusOK, res)
}
