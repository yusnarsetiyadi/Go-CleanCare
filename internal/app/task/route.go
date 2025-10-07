package task

import (
	"iss_cleancare/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.GET("", h.Find, middleware.Authentication)

	h.TaskTypeHandler.Route(v.Group("/type"))
}
