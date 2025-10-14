package model

import "cleancare/internal/abstraction"

type TaskTypeEntity struct {
	Name     string `json:"name"`
	TaskId   int    `json:"task_id"`
	IsDelete bool   `json:"is_delete"`
}

// TaskTypeEntityModel ...
type TaskTypeEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	TaskTypeEntity

	Task TaskEntityModel `json:"task" gorm:"foreignKey:TaskId"`

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (TaskTypeEntityModel) TableName() string {
	return "task_type"
}

type TaskTypeCountDataModel struct {
	Count int `json:"count"`
}
