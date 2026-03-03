package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITest godoc
// @Summary      Получить
// @Description  Удаляет запись по её ID (Soft Delete)
// @Tags         test
// @Produce      json
// @Param        id   path      string  true  "ID записи" format(string)
// @Success      200  {object}  TestDTO "Тестовые данные"
// @Failure      404  {object}  ErrorDTO "Запись не найдена"
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test/{id} [get]
func (cr *AppChiRouter) getAPITest(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("getAPITest start, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)
	defer cr.log.Debugf("getAPITest finish, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)

	res, err := cr.testFacade.Get(r.Context(), id)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	cr.renderJSON(rw, http.StatusOK, res)
}
