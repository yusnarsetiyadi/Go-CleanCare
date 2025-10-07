package dto

type TaskTypeCreateRequest struct {
	Name   string `json:"name" form:"name" validate:"required"`
	TaskId int    `json:"task_id" form:"task_id" validate:"required"`
}

type TaskTypeDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type TaskTypeUpdateRequest struct {
	ID     int     `param:"id" validate:"required"`
	Name   *string `json:"name" form:"name"`
	TaskId *int    `json:"task_id" form:"task_id"`
}
