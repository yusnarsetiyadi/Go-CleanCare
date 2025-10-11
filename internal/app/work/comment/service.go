package comment

import (
	"errors"
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/dto"
	"iss_cleancare/internal/factory"
	"iss_cleancare/internal/model"
	"iss_cleancare/internal/repository"
	"iss_cleancare/pkg/util/general"
	"iss_cleancare/pkg/util/response"
	"iss_cleancare/pkg/util/trxmanager"
	"net/http"

	"gorm.io/gorm"
)

type Service interface {
	FindByWorkId(ctx *abstraction.Context, payload *dto.CommentFindByWorkIdRequest) (map[string]interface{}, error)
	Create(ctx *abstraction.Context, payload *dto.CommentCreateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.CommentDeleteByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.CommentUpdateRequest) (map[string]interface{}, error)
}

type service struct {
	CommentRepository repository.Comment
	WorkRepository    repository.Work

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		CommentRepository: f.CommentRepository,
		WorkRepository:    f.WorkRepository,

		DB: f.Db,
	}
}

func (s *service) FindByWorkId(ctx *abstraction.Context, payload *dto.CommentFindByWorkIdRequest) (map[string]interface{}, error) {
	workData, err := s.WorkRepository.FindById(ctx, payload.WorkId)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if workData == nil {
		return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "work not found")
	}

	data, err := s.CommentRepository.FindByWorkId(ctx, payload.WorkId, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.CommentRepository.CountByWorkId(ctx, payload.WorkId)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	var res []map[string]interface{} = nil
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":         v.ID,
			"work_id":    v.WorkId,
			"comment":    v.Comment,
			"is_delete":  v.IsDelete,
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
			"created_by": map[string]interface{}{
				"id":   v.CreateBy.ID,
				"name": v.CreateBy.Name,
			},
			"updated_by": map[string]interface{}{
				"id":   v.UpdateBy.ID,
				"name": v.UpdateBy.Name,
			},
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.CommentCreateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		workData, err := s.WorkRepository.FindById(ctx, payload.WorkId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if workData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "work not found")
		}

		modelComment := &model.CommentEntityModel{
			Context: ctx,
			CommentEntity: model.CommentEntity{
				WorkId:   payload.WorkId,
				Comment:  payload.Comment,
				IsDelete: false,
			},
		}
		if err := s.CommentRepository.Create(ctx, modelComment).Error; err != nil {
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

func (s *service) Delete(ctx *abstraction.Context, payload *dto.CommentDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		commentData, err := s.CommentRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if commentData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "comment not found")
		}

		newCommentData := new(model.CommentEntityModel)
		newCommentData.Context = ctx
		newCommentData.ID = commentData.ID
		newCommentData.IsDelete = true
		if err = s.CommentRepository.Update(ctx, newCommentData).Error; err != nil {
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

func (s *service) Update(ctx *abstraction.Context, payload *dto.CommentUpdateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		commentData, err := s.CommentRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if commentData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "comment not found")
		}

		newCommentData := new(model.CommentEntityModel)
		newCommentData.Context = ctx
		newCommentData.ID = payload.ID
		if payload.Comment != nil {
			newCommentData.Comment = *payload.Comment
		}
		if err = s.CommentRepository.Update(ctx, newCommentData).Error; err != nil {
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
