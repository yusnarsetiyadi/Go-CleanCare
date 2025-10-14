package work

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/app/work/comment"
	"cleancare/internal/dto"
	"cleancare/internal/factory"
	"cleancare/pkg/util/response"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service

	CommentHandler comment.Handler
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),

		CommentHandler: *comment.NewHandler(f),
	}
}

func (h *handler) Create(c echo.Context) (err error) {
	payload := new(dto.WorkCreateRequest)

	if err = c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.Request().ParseMultipartForm(64 << 20); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "error bind multipart/form-data").SendError(c)
		}
		payload.ImageBefore = c.Request().MultipartForm.File["image_before"]
		payload.ImageAfter = c.Request().MultipartForm.File["image_after"]
	}

	data, err := h.service.Create(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) Delete(c echo.Context) (err error) {
	payload := new(dto.WorkDeleteByIDRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.Delete(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) Find(c echo.Context) (err error) {
	data, err := h.service.Find(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) FindById(c echo.Context) (err error) {
	payload := new(dto.WorkFindByIDRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.FindById(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) Update(c echo.Context) (err error) {
	payload := new(dto.WorkUpdateRequest)

	if err = c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.Request().ParseMultipartForm(64 << 20); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "error bind multipart/form-data").SendError(c)
		}
		payload.ImageBefore = c.Request().MultipartForm.File["image_before"]
		payload.ImageAfter = c.Request().MultipartForm.File["image_after"]
	}

	data, err := h.service.Update(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) Export(c echo.Context) (err error) {
	payload := new(dto.WorkExportRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	filename, data, format, err := h.service.Export(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SendBlobData(c, filename, *data, format)
}

func (h handler) DashboardAdmin(c echo.Context) (err error) {
	payload := new(dto.WorkDashboardAdminRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.DashboardAdmin(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) DashboardStaf(c echo.Context) (err error) {
	payload := new(dto.WorkDashboardStafRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.DashboardStaf(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}
