package rest

import (
	"net/http"

	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/go-chi/chi/v5"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITest обработчик пути GET /api/test/{id}
//
// @Summary      Получить
// @Description  Получает запись по её ID
// @Tags         test
// @Produce      json
// @Param        id   path      string  true  "ID записи" format(string)
// @Success      200  {object}  Test "Тестовые данные"
// @Failure      404  {object}  ErrorDTO "Запись не найдена"
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test/{id} [get]
func (cr *AppChiRouter) getAPITest(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("getAPITest start, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)
	defer cr.log.Debugf("getAPITest finish, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)

	res, err := cr.testFacade.Get(r.Context(), id)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
