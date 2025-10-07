package repository

import (
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/model"
	"iss_cleancare/pkg/util/general"

	"gorm.io/gorm"
)

type TaskType interface {
	FindById(ctx *abstraction.Context, id int) (*model.TaskTypeEntityModel, error)
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.TaskTypeEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
	Create(ctx *abstraction.Context, data *model.TaskTypeEntityModel) *gorm.DB
	Update(ctx *abstraction.Context, data *model.TaskTypeEntityModel) *gorm.DB
}

type task_type struct {
	abstraction.Repository
}

func NewTaskType(db *gorm.DB) *task_type {
	return &task_type{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *task_type) FindById(ctx *abstraction.Context, id int) (*model.TaskTypeEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.TaskTypeEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *task_type) Find(ctx *abstraction.Context, no_paging bool) (data []*model.TaskTypeEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "task_type", "is_delete = @false")
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Preload("Task").
		Find(&data).
		Error
	return
}

func (r *task_type) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "task_type", "is_delete = @false")
	var count model.TaskTypeCountDataModel
	err = r.CheckTrx(ctx).
		Table("task_type").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *task_type) Create(ctx *abstraction.Context, data *model.TaskTypeEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *task_type) Update(ctx *abstraction.Context, data *model.TaskTypeEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}
