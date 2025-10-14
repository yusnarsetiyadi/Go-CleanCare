package tipe

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/dto"
	"cleancare/internal/factory"
	"cleancare/internal/model"
	"cleancare/internal/repository"
	"cleancare/pkg/util/response"
	"cleancare/pkg/util/trxmanager"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type Service interface {
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	Create(ctx *abstraction.Context, payload *dto.TaskTypeCreateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.TaskTypeDeleteByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.TaskTypeUpdateRequest) (map[string]interface{}, error)
}

type service struct {
	TaskTypeRepository repository.TaskType
	TaskRepository     repository.Task

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		TaskTypeRepository: f.TaskTypeRepository,
		TaskRepository:     f.TaskRepository,

		DB: f.Db,
	}
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	data, err := s.TaskTypeRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.TaskTypeRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	var res []map[string]interface{} = nil
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":      v.ID,
			"name":    v.Name,
			"task_id": v.TaskId,
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.TaskTypeCreateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		taskData, err := s.TaskRepository.FindById(ctx, payload.TaskId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if taskData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task id not found")
		}

		modelTaskType := &model.TaskTypeEntityModel{
			Context: ctx,
			TaskTypeEntity: model.TaskTypeEntity{
				Name:     payload.Name,
				TaskId:   payload.TaskId,
				IsDelete: false,
			},
		}
		if err := s.TaskTypeRepository.Create(ctx, modelTaskType).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.TaskTypeDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		taskTypeData, err := s.TaskTypeRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if taskTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task type not found")
		}

		newTaskTypeData := new(model.TaskTypeEntityModel)
		newTaskTypeData.Context = ctx
		newTaskTypeData.ID = taskTypeData.ID
		newTaskTypeData.IsDelete = true

		if err = s.TaskTypeRepository.Update(ctx, newTaskTypeData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.TaskTypeUpdateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		taskTypeData, err := s.TaskTypeRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if taskTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task type not found")
		}

		newTaskTypeData := new(model.TaskTypeEntityModel)
		newTaskTypeData.Context = ctx
		newTaskTypeData.ID = payload.ID
		if payload.Name != nil {
			newTaskTypeData.Name = *payload.Name
		}
		if payload.TaskId != nil {
			taskData, err := s.TaskRepository.FindById(ctx, *payload.TaskId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if taskData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task not found")
			}
			newTaskTypeData.TaskId = *payload.TaskId
		}

		if err = s.TaskTypeRepository.Update(ctx, newTaskTypeData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success update!",
	}, nil
}
