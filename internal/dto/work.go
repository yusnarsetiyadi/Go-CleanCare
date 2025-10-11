package dto

import "mime/multipart"

type WorkCreateRequest struct {
	TaskId      int    `json:"task_id" form:"task_id" validate:"required"`
	TaskTypeId  int    `json:"task_type_id" form:"task_type_id" validate:"required"`
	Floor       string `json:"floor" form:"floor" validate:"required"`
	Info        string `json:"info" form:"info" validate:"required"`
	ImageBefore []*multipart.FileHeader
	ImageAfter  []*multipart.FileHeader
}

type WorkDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type WorkFindByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type WorkUpdateRequest struct {
	ID                int     `param:"id" validate:"required"`
	TaskId            *int    `json:"task_id" form:"task_id"`
	TaskTypeId        *int    `json:"task_type_id" form:"task_type_id"`
	Floor             *string `json:"floor" form:"floor"`
	Info              *string `json:"info" form:"info"`
	ImageBefore       []*multipart.FileHeader
	DeleteImageBefore *bool `json:"delete_image_before" form:"delete_image_before"`
	ImageAfter        []*multipart.FileHeader
	DeleteImageAfter  *bool `json:"delete_image_after" form:"delete_image_after"`
}

type WorkExportRequest struct {
	Format string `query:"format" validate:"required"`
}

type WorkDashboardAdminRequest struct {
	TaskId    int    `query:"task_id" validate:"required"`
	CreatedAt string `query:"created_at" validate:"required"`
}

type WorkDashboardStafRequest struct {
	TaskId    int    `query:"task_id" validate:"required"`
	CreatedAt string `query:"created_at" validate:"required"`
}
