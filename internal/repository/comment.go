package repository

import (
	"fmt"
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/model"
	"iss_cleancare/pkg/util/general"

	"gorm.io/gorm"
)

type Comment interface {
	FindByWorkId(ctx *abstraction.Context, workId int, no_paging bool) (data []*model.CommentEntityModel, err error)
	CountByWorkId(ctx *abstraction.Context, workId int) (data *int, err error)
	Create(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.CommentEntityModel, error)
	Update(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB
	FindByWorkIdArr(ctx *abstraction.Context, work_id int, no_paging bool) (data []*model.CommentEntityModel, err error)
}

type comment struct {
	abstraction.Repository
}

func NewComment(db *gorm.DB) *comment {
	return &comment{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *comment) FindByWorkId(ctx *abstraction.Context, workId int, no_paging bool) (data []*model.CommentEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "comment", "is_delete = @false"+fmt.Sprintf(" AND work_id = %d", workId))
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Preload("CreateBy").
		Preload("CreateBy.Role").
		Preload("UpdateBy").
		Preload("UpdateBy.Role").
		Find(&data).
		Error
	return
}

func (r *comment) CountByWorkId(ctx *abstraction.Context, workId int) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "comment", "is_delete = @false"+fmt.Sprintf(" AND work_id = %d", workId))
	var count model.CommentCountDataModel
	err = r.CheckTrx(ctx).
		Table("comment").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *comment) Create(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *comment) FindById(ctx *abstraction.Context, id int) (*model.CommentEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.CommentEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		Preload("Work").
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *comment) Update(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}

func (r *comment) FindByWorkIdArr(ctx *abstraction.Context, work_id int, no_paging bool) (data []*model.CommentEntityModel, err error) {
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where("work_id = ? AND is_delete = ?", work_id, false).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&data).
		Error
	return
}
