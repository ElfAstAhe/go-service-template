package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// deleteAPITest godoc
// @Summary      Удаление тестовых данных
// @Description  Удаляет запись по её ID
// @Tags         test
// @Param        id   path      string  true  "ID записи" format(string)
// @Success      204  "Запись успешно удалена, тело ответа отсутствует"
// @Failure      404  {object}  ErrorDTO "Запись не найдена"
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test/{id} [delete]
func (cr *AppChiRouter) deleteAPITest(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("deleteAPITest start, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)
	defer cr.log.Debugf("deleteAPITest finish, requestID [%s] path param [%s]", middleware.GetReqID(r.Context()), id)

	err := cr.testFacade.Delete(r.Context(), id)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	cr.renderEmpty(rw, http.StatusNoContent)
}
