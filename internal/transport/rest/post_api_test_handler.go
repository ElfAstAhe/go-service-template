package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

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
