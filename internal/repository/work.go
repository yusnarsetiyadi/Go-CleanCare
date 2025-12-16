package repository

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/model"
	"cleancare/pkg/util/general"
	"strings"

	"gorm.io/gorm"
)

type Work interface {
	Create(ctx *abstraction.Context, data *model.WorkEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.WorkEntityModel, error)
	Update(ctx *abstraction.Context, data *model.WorkEntityModel) *gorm.DB
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.WorkEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
	UpdateToNull(ctx *abstraction.Context, data *model.WorkEntityModel, column string) *gorm.DB
	FindByTaskIdArrAdmin(ctx *abstraction.Context, task_id int, created_at string, no_paging bool) (floorSummary []*model.FloorSummary, userSummary []*model.UserSummary, errFloor, errUser error)
	FindByTaskIdArrStaf(ctx *abstraction.Context, task_id int, created_at string, no_paging bool) (taskTypeSummary []*model.TaskTypeSummary, err error)
	FindByUserIdTaskIdTaskTypeIdFloor(ctx *abstraction.Context, userId, taskId, taskTypeId int, floor string) (*model.WorkEntityModel, error)
}

type work struct {
	abstraction.Repository
}

func NewWork(db *gorm.DB) *work {
	return &work{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *work) Create(ctx *abstraction.Context, data *model.WorkEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *work) FindById(ctx *abstraction.Context, id int) (*model.WorkEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.WorkEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		Preload("User").
		Preload("Task").
		Preload("TaskType").
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *work) Update(ctx *abstraction.Context, data *model.WorkEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}

func (r *work) Find(ctx *abstraction.Context, no_paging bool) (data []*model.WorkEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "work", "is_delete = @false")
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Preload("User").
		Preload("Task").
		Preload("TaskType").
		Find(&data).
		Error
	return
}

func (r *work) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "work", "is_delete = @false")
	var count model.WorkCountDataModel
	err = r.CheckTrx(ctx).
		Table("work").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *work) UpdateToNull(ctx *abstraction.Context, data *model.WorkEntityModel, column string) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Update(column, nil)
}

func (r *work) FindByTaskIdArrAdmin(ctx *abstraction.Context, task_id int, created_at string, no_paging bool) (floorSummary []*model.FloorSummary, userSummary []*model.UserSummary, errFloor, errUser error) {
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	valDate := strings.Split(created_at, "_")
	startDate := valDate[0] + " 00:00:00"
	endDate := valDate[1] + " 23:59:59"

	errFloor = r.CheckTrx(ctx).
		Model(&model.WorkEntityModel{}).
		Select("floor, COUNT(*) as count").
		Where("task_id = ? AND is_delete = ? AND created_at BETWEEN ? AND ?", task_id, false, startDate, endDate).
		Group("floor").
		Order("CAST(SUBSTRING(floor, 8, LENGTH(floor)) AS UNSIGNED) ASC").
		Limit(limit).
		Offset(offset).
		Scan(&floorSummary).Error

	errUser = r.CheckTrx(ctx).
		Model(&model.WorkEntityModel{}).
		Joins("JOIN user ON user.id = work.user_id").
		Select("work.user_id, user.name, COUNT(*) as count").
		Where("work.task_id = ? AND work.is_delete = ? AND work.created_at BETWEEN ? AND ?", task_id, false, startDate, endDate).
		Group("work.user_id, user.name").
		Order("work.user_id ASC").
		Limit(limit).
		Offset(offset).
		Scan(&userSummary).Error

	return
}

func (r *work) FindByTaskIdArrStaf(ctx *abstraction.Context, task_id int, created_at string, no_paging bool) (taskTypeSummary []*model.TaskTypeSummary, err error) {
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	valDate := strings.Split(created_at, "_")
	startDate := valDate[0] + " 00:00:00"
	endDate := valDate[1] + " 23:59:59"

	err = r.CheckTrx(ctx).
		Model(&model.WorkEntityModel{}).
		Joins("JOIN task_type ON task_type.id = work.task_type_id").
		Select("work.task_type_id, task_type.name, COUNT(*) as count").
		Where("work.task_id = ? AND work.is_delete = ? AND work.created_at BETWEEN ? AND ?", task_id, false, startDate, endDate).
		Group("work.task_type_id, task_type.name").
		Order("work.task_type_id ASC").
		Limit(limit).
		Offset(offset).
		Scan(&taskTypeSummary).Error

	return
}

func (r *work) FindByUserIdTaskIdTaskTypeIdFloor(ctx *abstraction.Context, userId, taskId, taskTypeId int, floor string) (*model.WorkEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.WorkEntityModel
	err := conn.
		Where("user_id = ? AND task_id = ? AND task_type_id = ? AND floor = ? AND DATE(created_at) = CURRENT_DATE AND is_delete = ?", userId, taskId, taskTypeId, floor, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}
