package rest

import (
	"net/http"

	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/go-chi/chi/v5"

	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// deleteAPITest обработчик пути DELETE /api/test/{id}
//
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

	cr.log.Debugf("deleteAPITest start, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)
	defer cr.log.Debugf("deleteAPITest finish, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)

	err := cr.testFacade.Delete(r.Context(), id)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderEmpty(rw, http.StatusNoContent)
}
