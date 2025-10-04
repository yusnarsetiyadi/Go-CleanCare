package test

import (
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/dto"
	"iss_cleancare/internal/factory"
	"iss_cleancare/pkg/util/response"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service
}

var err error

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

// @Summary      Test
// @Description  Test
// @Tags         Test
// @Accept       json
// @Produce      json
// @Success      200      {object}  dto.TestResponse
// @Failure      400      {object}  res.errorResponse
// @Failure      401      {object}  res.errorResponse
// @Failure      404      {object}  res.errorResponse
// @Failure      500      {object}  res.errorResponse
// @Router       /api/v1/test [get]
func (h *handler) Test(c echo.Context) error {
	cc := c.(*abstraction.Context)

	payload := new(dto.TestResponse)
	if err = c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}

	data, err := h.service.Test(cc)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}

	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) TestGomail(c echo.Context) error {
	cc := c.(*abstraction.Context)

	payload := new(dto.TestGomailRequest)
	if err = c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}

	data, err := h.service.TestGomail(cc, payload.Recipient)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}

	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) TestDriveCreate(c echo.Context) error {
	cc := c.(*abstraction.Context)

	if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}

	formFiles := c.Request().MultipartForm.File["files"]
	var files []*multipart.FileHeader
	files = append(files, formFiles...)

	data, err := h.service.TestDriveCreate(cc, files)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}

	return response.SuccessResponse(data).SendSuccess(c)
}

func (h *handler) TestDriveGetById(c echo.Context) error {
	cc := c.(*abstraction.Context)

	data, err := h.service.TestDriveGetById(cc, cc.Param("id"))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}

	return response.SuccessResponse(data).SendSuccess(c)
}
