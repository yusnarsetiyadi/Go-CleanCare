package model

import (
	"iss_cleancare/internal/abstraction"

	"gorm.io/gorm"
)

type CommentEntity struct {
	WorkId   int    `json:"work_id"`
	Comment  string `json:"comment"`
	IsDelete bool   `json:"is_delete"`
}

// CommentEntityModel ...
type CommentEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	CommentEntity

	abstraction.EntityWithBy

	Work     WorkEntityModel `json:"work" gorm:"foreignKey:WorkId"`
	CreateBy UserEntityModel `json:"create_by" gorm:"foreignKey:CreatedBy"`
	UpdateBy UserEntityModel `json:"update_by" gorm:"foreignKey:UpdatedBy"`

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (CommentEntityModel) TableName() string {
	return "comment"
}

type CommentCountDataModel struct {
	Count int `json:"count"`
}

func (m *CommentEntityModel) BeforeUpdate(tx *gorm.DB) (err error) {
	// m.UpdatedAt = general.NowLocal()
	m.UpdatedBy = &m.Context.Auth.ID
	return
}

func (m *CommentEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	m.CreatedBy = m.Context.Auth.ID
	return
}
