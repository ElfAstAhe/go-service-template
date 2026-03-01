package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
	"github.com/go-chi/chi/v5/middleware"
)

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
