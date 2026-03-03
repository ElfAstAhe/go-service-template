package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// postAPITest godoc
// @Summary      Создание новых тестовых данных
// @Description  Сохраняет новые тестовые данные
// @Tags         test
// @Accept       json
// @Produce      json
// @Param        input  body      TestDTO  true  "Тестовые данные"
// @Success      201    {object}  TestDTO
// @Failure      400    {object}  ErrorDTO
// @Failure      409    {object}  ErrorDTO
// @Failure      500    "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test [post]
func (cr *AppChiRouter) postAPITest(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("postAPITest start, requestID [%s]", middleware.GetReqID(r.Context()))
	defer cr.log.Debugf("postAPITest finish, requestID [%s]", middleware.GetReqID(r.Context()))

	var income = &dto.TestDTO{}
	err := cr.decodeJSON(r, income)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	res, err := cr.testFacade.Create(r.Context(), income)
	if err != nil {
		cr.renderError(rw, err)

		return
	}
	location := r.URL.JoinPath(res.ID)
	rw.Header().Set("Location", location.String())

	cr.renderJSON(rw, http.StatusCreated, res)
}
