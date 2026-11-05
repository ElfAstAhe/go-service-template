package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// getAPITest2List обработчик пути GET /api/test2
//
// @Summary      Получить
// @Description  Получить список данных 2
// @Tags         test
// @Produce      json
// @Param        limit   query   int  false  "limit row count, max 1000" format(int)
// @Param        offset  query   int  false  "offset, min 0, max n" format(int)
// @Success      200  {array}  Test2 "Набор тестовых данных 2"
// @Failure      400  {object} ErrorDTO
// @Failure      500  "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test2 [get]
func (cr *AppChiRouter) getAPITest2List(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("getAPITest2List start, requestID [%s]", pkgtr.RequestID(r.Context()))
	defer cr.log.Debugf("getAPITest2List finish, requestID [%s]", pkgtr.RequestID(r.Context()))

	limit := pkghttp.GetQueryIntDefault(r, "limit", transport.DefaultListLimit)
	offset := pkghttp.GetQueryIntDefault(r, "offset", transport.DefaultListOffset)

	res, err := cr.test2Facade.List(r.Context(), limit, offset)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	pkghttp.RenderJSONDefault(rw, http.StatusOK, res)
}
