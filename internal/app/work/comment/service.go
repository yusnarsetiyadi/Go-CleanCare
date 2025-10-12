package comment

import (
	"errors"
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/dto"
	"iss_cleancare/internal/factory"
	"iss_cleancare/internal/model"
	"iss_cleancare/internal/repository"
	"iss_cleancare/pkg/constant"
	"iss_cleancare/pkg/util/general"
	"iss_cleancare/pkg/util/response"
	"iss_cleancare/pkg/util/trxmanager"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-redis/redis/v8"
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
	UserRepository    repository.User

	DB      *gorm.DB
	DbRedis *redis.Client
}

func NewService(f *factory.Factory) Service {
	return &service{
		CommentRepository: f.CommentRepository,
		WorkRepository:    f.WorkRepository,
		UserRepository:    f.UserRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
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
		userDataUnreadComment := general.GetUserIdArrayFromKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(v.ID))
		if slices.Contains(userDataUnreadComment, strconv.Itoa(ctx.Auth.ID)) {
			general.RemoveUserIdFromKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(v.ID), ctx.Auth.ID)
		}

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

		userAdmin, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_ADMIN, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if ctx.Auth.RoleID == constant.ROLE_ID_STAFF {
			for _, v := range userAdmin {
				general.AppendUserIdToKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(modelComment.ID), v.ID)
			}
		} else {
			general.AppendUserIdToKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(modelComment.ID), workData.UserId)
			for _, v := range userAdmin {
				if v.ID != ctx.Auth.ID {
					general.AppendUserIdToKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(modelComment.ID), v.ID)
				}
			}
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

		sendUnread := false
		newCommentData := new(model.CommentEntityModel)
		newCommentData.Context = ctx
		newCommentData.ID = payload.ID
		if payload.Comment != nil {
			sendUnread = true
			newCommentData.Comment = *payload.Comment
		}
		if err = s.CommentRepository.Update(ctx, newCommentData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if sendUnread {
			userAdmin, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_ADMIN, true)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if ctx.Auth.RoleID == constant.ROLE_ID_STAFF {
				for _, v := range userAdmin {
					general.AppendUserIdToKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(newCommentData.ID), v.ID)
				}
			} else {
				general.AppendUserIdToKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(newCommentData.ID), commentData.Work.UserId)
				for _, v := range userAdmin {
					if v.ID != ctx.Auth.ID {
						general.AppendUserIdToKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(newCommentData.ID), v.ID)
					}
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success update!",
	}, nil
}
