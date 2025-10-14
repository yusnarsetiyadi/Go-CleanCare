package auth

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/dto"
	"cleancare/internal/factory"
	"cleancare/pkg/util/general"
	"cleancare/pkg/util/response"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

func (h *handler) Login(c echo.Context) error {
	payload := new(dto.AuthLoginRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err := c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.Login(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) Logout(c echo.Context) error {
	data, err := h.service.Logout(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) RefreshToken(c echo.Context) error {
	data, err := h.service.RefreshToken(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) SendEmailForgotPassword(c echo.Context) error {
	payload := new(dto.AuthSendEmailForgotPasswordRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err := c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.SendEmailForgotPassword(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) ValidationResetPassword(c echo.Context) error {
	payload := new(dto.AuthValidationResetPasswordRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err := c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.ValidationResetPassword(c.(*abstraction.Context), payload)
	if err != nil {
		htmlContent := general.ProcessHTMLResponseEmail("assets/html/webview/reset_password_failed.html", "{{.Error}}", err.Error())
		return c.HTML(200, htmlContent)
	}
	htmlContent := general.ProcessHTMLResponseEmail("assets/html/webview/reset_password_success.html", "{{.Data}}", data)
	return c.HTML(200, htmlContent)
}

func (h *handler) VerifyNumber(c echo.Context) error {
	payload := new(dto.AuthVerifyNumberRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err := c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.VerifyNumber(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) Register(c echo.Context) error {
	payload := new(dto.AuthRegisterRequest)

	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err := c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.Request().ParseMultipartForm(64 << 20); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "error bind multipart/form-data").SendError(c)
		}
		payload.Profile = c.Request().MultipartForm.File["profile"]
	}

	data, err := h.service.Register(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}
