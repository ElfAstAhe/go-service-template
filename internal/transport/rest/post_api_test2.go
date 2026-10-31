package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	pkgtr "github.com/ElfAstAhe/go-service-template/pkg/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"

	_ "github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	_ "github.com/ElfAstAhe/go-service-template/internal/transport"
)

// postAPITest2 обработчик пути POST /api/test2
//
// @Summary      Создание новых тестовых данных
// @Description  Сохраняет новые тестовые данные 2
// @Tags         test2
// @Accept       json
// @Produce      json
// @Param        input  body      Test2  true  "Тестовые данные"
// @Success      201    {object}  Test2
// @Failure      400    {object}  ErrorDTO
// @Failure      409    {object}  ErrorDTO
// @Failure      500    "Внутренняя ошибка сервера (пустое тело)"
// @Router       /api/test2 [post]
//
//goland:noinspection DuplicatedCode
func (cr *AppChiRouter) postAPITest2(rw http.ResponseWriter, r *http.Request) {
	cr.log.Debugf("postAPITest2 start, requestID [%s]", pkgtr.RequestID(r.Context()))
	defer cr.log.Debugf("postAPITest2 finish, requestID [%s]", pkgtr.RequestID(r.Context()))

	var income = &dto.Test2{}
	err := pkghttp.DecodeJSON(r, income)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}

	res, err := cr.test2Facade.Create(r.Context(), income)
	if err != nil {
		pkghttp.RenderErrorDefault(rw, err)

		return
	}
	location := r.URL.JoinPath(res.ID)
	rw.Header().Set("Location", location.String())

	pkghttp.RenderJSONDefault(rw, http.StatusCreated, res)
}
