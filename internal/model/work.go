package model

import (
	"cleancare/internal/abstraction"

	"gorm.io/gorm"
)

type WorkEntity struct {
	UserId      int     `json:"user_id"`
	TaskId      int     `json:"task_id"`
	TaskTypeId  int     `json:"task_type_id"`
	Floor       string  `json:"floor"`
	Info        string  `json:"info"`
	ImageBefore *string `json:"image_before"`
	ImageAfter  *string `json:"image_after"`
	IsDelete    bool    `json:"is_delete"`
}

// WorkEntityModel ...
type WorkEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	WorkEntity

	abstraction.Entity

	User     UserEntityModel     `json:"user" gorm:"foreignKey:UserId"`
	Task     TaskEntityModel     `json:"task" gorm:"foreignKey:TaskId"`
	TaskType TaskTypeEntityModel `json:"task_type" gorm:"foreignKey:TaskTypeId"`

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (WorkEntityModel) TableName() string {
	return "work"
}

type WorkCountDataModel struct {
	Count int `json:"count"`
}

func (m *WorkEntityModel) BeforeUpdate(tx *gorm.DB) (err error) {
	// m.UpdatedAt = general.NowLocal()
	return
}

func (m *WorkEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	return
}

type FloorSummary struct {
	Floor string `json:"floor"`
	Count int    `json:"count"`
}

type UserSummary struct {
	UserId int    `json:"user_id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
}

type TaskTypeSummary struct {
	TaskTypeId int    `json:"task_type_id"`
	Name       string `json:"name"`
	Count      int    `json:"count"`
}
