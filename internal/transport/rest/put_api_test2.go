package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/go-chi/chi/v5"

	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// putAPITest2 обработчик пути PUT /api/test2/{id}
//
// @Summary      Изменение тестовых данных
// @Description  Изменяет тестовые данные
// @Tags         test
// @Accept       json
// @Produce      json
// @Param        id     path      string   true  "ID записи" format(string)
// @Param        input  body      Test2  true  "Тестовые данные"
// @Success      200    {object}  Test2
// @Failure      400    {object}  ErrorDTO
// @Failure      404    {object}  ErrorDTO
// @Failure      409    {object}  ErrorDTO
// @Failure      500    "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test2/{id} [put]
//
//goland:noinspection DuplicatedCode
func (cr *AppChiRouter) putAPITest2(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cr.log.Debugf("putAPITest2 start, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)
	defer cr.log.Debugf("putAPITest2 finish, requestID [%s] path param [%s]", pkgtr.RequestID(r.Context()), id)

	var income = &dto.Test2{}
	err := pkghttp.DecodeJSON(r, income)
	if err != nil {
		pkghttp.RenderError(rw, err, pkghttp.MapToHTTPStatus)

		return
	}

	res, err := cr.test2Facade.Change(r.Context(), id, income)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
