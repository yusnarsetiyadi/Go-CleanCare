package repository

import (
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/model"
	"iss_cleancare/pkg/util/general"

	"gorm.io/gorm"
)

type User interface {
	FindByNumberId(ctx *abstraction.Context, numberId string) (*model.UserEntityModel, error)
	FindByEmail(ctx *abstraction.Context, email string) (*model.UserEntityModel, error)
	Create(ctx *abstraction.Context, data *model.UserEntityModel) *gorm.DB
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.UserEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
	FindById(ctx *abstraction.Context, id int) (*model.UserEntityModel, error)
	Update(ctx *abstraction.Context, data *model.UserEntityModel) *gorm.DB
	UpdateToNull(ctx *abstraction.Context, data *model.UserEntityModel, column string) *gorm.DB
}

type user struct {
	abstraction.Repository
}

func NewUser(db *gorm.DB) *user {
	return &user{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *user) FindByNumberId(ctx *abstraction.Context, numberId string) (*model.UserEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.UserEntityModel
	err := conn.
		Where("LOWER(number_id) = LOWER(?) AND is_delete = ?", numberId, false).
		Preload("Role").
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *user) FindByEmail(ctx *abstraction.Context, email string) (*model.UserEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.UserEntityModel
	err := conn.
		Where("LOWER(email) = LOWER(?) AND is_delete = ?", email, false).
		Preload("Role").
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *user) Create(ctx *abstraction.Context, data *model.UserEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *user) Find(ctx *abstraction.Context, no_paging bool) (data []*model.UserEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "user", "is_delete = @false")
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Preload("Role").
		Find(&data).
		Error
	return
}

func (r *user) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "user", "is_delete = @false")
	var count model.UserCountDataModel
	err = r.CheckTrx(ctx).
		Table("user").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *user) FindById(ctx *abstraction.Context, id int) (*model.UserEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.UserEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		Preload("Role").
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *user) Update(ctx *abstraction.Context, data *model.UserEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}

func (r *user) UpdateToNull(ctx *abstraction.Context, data *model.UserEntityModel, column string) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Update(column, nil)
}
