package dto

type CommentFindByWorkIdRequest struct {
	WorkId int `param:"work_id" validate:"required"`
}

type CommentCreateRequest struct {
	Comment string `json:"comment" form:"comment" validate:"required"`
	WorkId  int    `json:"work_id" form:"work_id" validate:"required"`
}

type CommentDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type CommentUpdateRequest struct {
	ID      int     `param:"id" validate:"required"`
	Comment *string `json:"comment" form:"comment"`
}
