package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/go-chi/chi/v5/middleware"
)

func (cr *AppChiRouter) getAPITestSearch(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("getAPITestSearch start, requestID [%s]", middleware.GetReqID(r.Context()))
	defer cr.log.Debugf("getAPITestSearch finish, requestID [%s]", middleware.GetReqID(r.Context()))

	code := cr.getQueryString(r, "code", "")
	if code == "" {
		cr.renderError(rw, errs.NewInvalidArgumentError("code", ""))

		return
	}

	res, err := cr.testFacade.GetByCode(r.Context(), code)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	cr.renderJSON(rw, http.StatusOK, res)
}
