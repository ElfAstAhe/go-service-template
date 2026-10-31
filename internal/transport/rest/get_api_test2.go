package rest

import (
	"net/http"

	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/go-chi/chi/v5"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITest2 обработчик пути GET /api/test2/{id}
//
// @Summary      Получить
// @Description  Удаляет запись по её ID (Soft Delete)
// @Tags         test2
// @Produce      json
// @Param        id   path      string  true  "ID записи" format(string)
// @Success      200  {object}  Test2 "Тестовые данные 2"
// @Failure      404  {object}  ErrorDTO "Запись не найдена"
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test2/{id} [get]
func (cr *AppChiRouter) getAPITest2(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("getAPITest2 start, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)
	defer cr.log.Debugf("getAPITest2 finish, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)

	res, err := cr.test2Facade.Get(r.Context(), id)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
