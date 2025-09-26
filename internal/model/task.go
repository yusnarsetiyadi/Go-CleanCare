package model

import "iss_cleancare/internal/abstraction"

type TaskEntity struct {
	Name     string `json:"name"`
	IsDelete bool   `json:"is_delete"`
}

// TaskEntityModel ...
type TaskEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	TaskEntity

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (TaskEntityModel) TableName() string {
	return "task"
}

type TaskCountDataModel struct {
	Count int `json:"count"`
}
