package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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
