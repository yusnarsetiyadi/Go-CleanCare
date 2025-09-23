package contact

import (
	"iss_cleancare/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.POST("", h.Create, middleware.Authentication)
	v.GET("", h.Find, middleware.Authentication)
	v.GET("/:id", h.FindById, middleware.Authentication)
	v.PUT("/:id", h.Update, middleware.Authentication)
	v.DELETE("/:id", h.Delete, middleware.Authentication)
	v.PATCH("/change-password/:id", h.ChangePassword, middleware.Authentication)
	v.PATCH("/reset-password/:id", h.ResetPassword, middleware.Authentication)
	v.GET("/info", h.GetUserInfo, middleware.Authentication)
	v.GET("/export", h.Export, middleware.Authentication)
}
