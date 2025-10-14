package repository

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/model"
	"cleancare/pkg/util/general"

	"gorm.io/gorm"
)

type Task interface {
	FindById(ctx *abstraction.Context, id int) (*model.TaskEntityModel, error)
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.TaskEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
}

type task struct {
	abstraction.Repository
}

func NewTask(db *gorm.DB) *task {
	return &task{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *task) FindById(ctx *abstraction.Context, id int) (*model.TaskEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.TaskEntityModel
	err := conn.
		Where("id = ?", id).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *task) Find(ctx *abstraction.Context, no_paging bool) (data []*model.TaskEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "task", "")
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&data).
		Error
	return
}

func (r *task) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "task", "")
	var count model.TaskCountDataModel
	err = r.CheckTrx(ctx).
		Table("task").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}
