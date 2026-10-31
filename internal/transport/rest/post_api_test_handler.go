package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"

	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// postAPITest обработчик пути POST /api/test
//
// @Summary      Создание новых тестовых данных
// @Description  Сохраняет новые тестовые данные
// @Tags         test
// @Accept       json
// @Produce      json
// @Param        input  body      Test  true  "Тестовые данные"
// @Success      201    {object}  Test
// @Failure      400    {object}  ErrorDTO
// @Failure      409    {object}  ErrorDTO
// @Failure      500    "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test [post]
func (cr *AppChiRouter) postAPITest(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("postAPITest start, requestID [%s]", pkgtr.RequestID(r.Context()))
	defer cr.log.Debugf("postAPITest finish, requestID [%s]", pkgtr.RequestID(r.Context()))

	var income = &dto.Test{}
	err := pkghttp.DecodeJSON(r, income)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	res, err := cr.testFacade.Create(r.Context(), income)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}
	location := r.URL.JoinPath(res.ID)
	rw.Header().Set("Location", location.String())

	pkghttp.RenderJSONDefault(rw, http.StatusCreated, res)
}
