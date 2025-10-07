package http

import (
	"fmt"
	"net/http"

	"iss_cleancare/internal/app/auth"
	"iss_cleancare/internal/app/role"
	"iss_cleancare/internal/app/task"
	"iss_cleancare/internal/app/test"
	"iss_cleancare/internal/app/user"
	"iss_cleancare/internal/config"
	"iss_cleancare/internal/factory"
	"iss_cleancare/pkg/constant"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Init(e *echo.Echo, f *factory.Factory) {

	e.GET("/", func(c echo.Context) error {
		message := fmt.Sprintf("Hello there, welcome to app %s version %s.", config.Get().App.App, config.Get().App.Version)
		return c.String(http.StatusOK, message)
	})

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Static("/images", constant.PATH_ASSETS_IMAGES)
	e.Static("/share", constant.PATH_SHARE)
	e.Static("/file_saved", constant.PATH_FILE_SAVED)

	test.NewHandler(f).Route(e.Group("/test"))
	auth.NewHandler(f).Route(e.Group("/auth"))
	role.NewHandler(f).Route(e.Group("/role"))
	task.NewHandler(f).Route(e.Group("/task"))
	user.NewHandler(f).Route(e.Group("/user"))
}
