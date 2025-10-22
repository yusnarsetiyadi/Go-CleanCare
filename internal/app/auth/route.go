package auth

import (
	"cleancare/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.POST("/login", h.Login)
	v.POST("/logout", h.Logout, middleware.Logout)
	v.POST("/refresh-token", h.RefreshToken, middleware.RefreshToken)
	v.POST("/send-email/forgot-password", h.SendEmailForgotPassword, middleware.ResetPasswordIpCheck)
	v.GET("/validation/reset-password/:token", h.ValidationResetPassword)
	v.POST("/verify-number", h.VerifyNumber, middleware.VerifyNumberIpCheck)
	v.POST("/register", h.Register, middleware.RegisterIpCheck)
}
