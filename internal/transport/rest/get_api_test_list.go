package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITestList godoc
// @Summary      Получить
// @Description  Удаляет запись по её ID (Soft Delete)
// @Tags         test
// @Produce      json
// @Param        limit   query   int  false  "limit row count, max 1000" format(int)
// @Param        offset  query   int  false  "offset, min 0, max n" format(int)
// @Success      200  {array}  TestDTO "Набор тестовых данных"
// @Failure      400  {object} ErrorDTO
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test [get]
func (cr *AppChiRouter) getAPITestList(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("getAPITestList start, requestID [%s]", middleware.GetReqID(r.Context()))
	defer cr.log.Debugf("getAPITestList finish, requestID [%s]", middleware.GetReqID(r.Context()))

	limit, err := cr.getQueryInt(r, "limit", transport.DefaultListLimit)
	if err != nil {
		cr.renderError(rw, err)

		return
	}
	offset, err := cr.getQueryInt(r, "offset", transport.DefaultListOffset)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	res, err := cr.testFacade.List(r.Context(), limit, offset)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	cr.renderJSON(rw, http.StatusOK, res)
}
