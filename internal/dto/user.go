package dto

import "mime/multipart"

type UserCreateRequest struct {
	Name     string `json:"name" form:"name" validate:"required"`
	NumberId string `json:"number_id" form:"number_id" validate:"required"`
	RoleId   int    `json:"role_id" form:"role_id" validate:"required"`
}

type UserFindByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type UserUpdateRequest struct {
	ID            int     `param:"id" validate:"required"`
	NumberId      *string `json:"number_id" form:"number_id"`
	Name          *string `json:"name" form:"name"`
	Email         *string `json:"email" form:"email"`
	RoleId        *int    `json:"role_id" form:"role_id"`
	Profile       []*multipart.FileHeader
	DeleteProfile *bool `json:"delete_profile" form:"delete_profile"`
}

type UserDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type UserChangePasswordRequest struct {
	ID          int    `param:"id" validate:"required"`
	OldPassword string `json:"old_password" form:"old_password" validate:"required"`
	NewPassword string `json:"new_password" form:"new_password" validate:"required"`
}
