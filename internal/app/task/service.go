package task

import (
	"cleancare/internal/abstraction"
	"cleancare/internal/factory"
	"cleancare/internal/repository"
	"cleancare/pkg/util/response"
	"net/http"

	"gorm.io/gorm"
)

type Service interface {
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
}

type service struct {
	TaskRepository repository.Task

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		TaskRepository: f.TaskRepository,

		DB: f.Db,
	}
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	data, err := s.TaskRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.TaskRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	var res []map[string]interface{} = nil
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":   v.ID,
			"name": v.Name,
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}
