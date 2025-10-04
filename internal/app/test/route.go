package test

import "github.com/labstack/echo/v4"

func (h *handler) Route(g *echo.Group) {
	g.GET("", h.Test)
	g.POST("/gomail", h.TestGomail)
	g.POST("/gdrive", h.TestDriveCreate)
	g.GET("/gdrive/:id", h.TestDriveGetById)
}
