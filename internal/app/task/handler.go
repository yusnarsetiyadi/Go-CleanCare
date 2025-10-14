package task

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/app/task/tipe"
	"cleancare/internal/factory"
	"cleancare/pkg/util/response"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service

	TaskTypeHandler tipe.Handler
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),

		TaskTypeHandler: *tipe.NewHandler(f),
	}
}

func (h handler) Find(c echo.Context) (err error) {
	data, err := h.service.Find(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}
