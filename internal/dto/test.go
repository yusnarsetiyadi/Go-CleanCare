package dto

import "mime/multipart"

type TestResponse struct {
	Message string `json:"message"`
}

type TestGomailRequest struct {
	Recipient string `json:"recipient"`
}

type TestDriveRequest struct {
	File multipart.FileHeader `json:"file" form:"file"`
}
