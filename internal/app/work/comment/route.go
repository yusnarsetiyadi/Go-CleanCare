package comment

import (
	"iss_cleancare/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Route(v *echo.Group) {
	v.GET("/:work_id", h.FindByWorkId, middleware.Authentication)
	v.POST("", h.Create, middleware.Authentication)
	v.DELETE("/:id", h.Delete, middleware.Authentication)
	v.PUT("/:id", h.Update, middleware.Authentication)
}
