package repository

import (
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/model"
	"iss_cleancare/pkg/util/general"

	"gorm.io/gorm"
)

type Role interface {
	FindById(ctx *abstraction.Context, id int) (*model.RoleEntityModel, error)
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.RoleEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
}

type role struct {
	abstraction.Repository
}

func NewRole(db *gorm.DB) *role {
	return &role{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *role) FindById(ctx *abstraction.Context, id int) (*model.RoleEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.RoleEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *role) Find(ctx *abstraction.Context, no_paging bool) (data []*model.RoleEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "role", "is_delete = @false")
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

func (r *role) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "role", "is_delete = @false")
	var count model.RoleCountDataModel
	err = r.CheckTrx(ctx).
		Table("role").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}
