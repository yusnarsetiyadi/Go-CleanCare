package work

import (
	"iss_cleancare/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.POST("", h.Create, middleware.Authentication)
	v.DELETE("/:id", h.Delete, middleware.Authentication)
	v.GET("", h.Find, middleware.Authentication)
	v.GET("/:id", h.FindById, middleware.Authentication)
	v.PUT("/:id", h.Update, middleware.Authentication)
	v.GET("/export", h.Export, middleware.Authentication)
	v.GET("/dashboard-admin", h.DashboardAdmin, middleware.Authentication)
	v.GET("/dashboard-staf", h.DashboardStaf, middleware.Authentication)

	h.CommentHandler.Route(v.Group("/comment"))
}
