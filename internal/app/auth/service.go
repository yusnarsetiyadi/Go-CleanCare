package auth

import (
	"context"
	"errors"
	"fmt"
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/config"
	"iss_cleancare/internal/dto"
	"iss_cleancare/internal/factory"
	"iss_cleancare/internal/model"
	modelToken "iss_cleancare/internal/model/token"
	"iss_cleancare/internal/repository"
	"iss_cleancare/pkg/constant"
	"iss_cleancare/pkg/gomail"
	"iss_cleancare/pkg/util/aescrypt"
	"iss_cleancare/pkg/util/encoding"
	"iss_cleancare/pkg/util/general"
	"iss_cleancare/pkg/util/response"
	"iss_cleancare/pkg/util/trxmanager"
	"iss_cleancare/pkg/ws"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Login(ctx *abstraction.Context, payload *dto.AuthLoginRequest) (map[string]interface{}, error)
	Logout(ctx *abstraction.Context) (map[string]interface{}, error)
	RefreshToken(ctx *abstraction.Context) (map[string]interface{}, error)
	SendEmailForgotPassword(ctx *abstraction.Context, payload *dto.AuthSendEmailForgotPasswordRequest) (map[string]interface{}, error)
	ValidationResetPassword(ctx *abstraction.Context, payload *dto.AuthValidationResetPasswordRequest) (string, error)
}

type service struct {
	UserRepository repository.User

	DB      *gorm.DB
	DbRedis *redis.Client
}

func NewService(f *factory.Factory) Service {
	return &service{
		UserRepository: f.UserRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
	}
}

func (s *service) encryptTokenClaims(v int) (encryptedString string, err error) {
	encryptedString, err = aescrypt.EncryptAES(fmt.Sprint(v), config.Get().JWT.SecretKey)
	return
}

func (s *service) Login(ctx *abstraction.Context, payload *dto.AuthLoginRequest) (map[string]interface{}, error) {
	var (
		err   error
		data  = new(model.UserEntityModel)
		token string
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		data, err = s.UserRepository.FindByNumberId(ctx, payload.NumberId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if data == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "number id or password is incorrect")
		}

		if err = bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(payload.Password)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "number id or password is incorrect")
		}

		var encryptedUserID string
		if encryptedUserID, err = s.encryptTokenClaims(data.ID); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		var encryptedUserRoleID string
		if encryptedUserRoleID, err = s.encryptTokenClaims(data.RoleId); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		encodedEmail := encoding.Encode(data.Email)
		uuidUserLogin := uuid.NewString()
		encodedUuidLogin := encoding.Encode(uuidUserLogin)

		tokenClaims := &modelToken.TokenClaims{
			ID:        encryptedUserID,
			RoleID:    encryptedUserRoleID,
			Email:     encodedEmail,
			UuidLogin: encodedUuidLogin,
			Exp:       time.Now().Add(time.Duration(24 * time.Hour)).Unix(),
		}
		authToken := modelToken.NewAuthToken(tokenClaims)
		token, err = authToken.Token()
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		general.AppendUUIDToRedisArray(s.DbRedis, general.GenerateRedisKeyUserLogin(data.ID), uuidUserLogin)

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token": token,
		"data": map[string]interface{}{
			"id":         data.ID,
			"number_id":  data.NumberId,
			"name":       data.Name,
			"email":      data.Email,
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		},
	}, nil
}

func (s *service) Logout(ctx *abstraction.Context) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {

		general.RemoveUUIDFromRedisArray(s.DbRedis, general.GenerateRedisKeyUserLogin(ctx.Auth.ID), ctx.Auth.UuidLogin)
		general.RemoveUUIDFromRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, ctx.Auth.UuidLogin)

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success logout!",
	}, nil
}

func (s *service) RefreshToken(ctx *abstraction.Context) (map[string]interface{}, error) {
	var token string
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		data, err := s.UserRepository.FindById(ctx, ctx.Auth.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		var encryptedUserID string
		if encryptedUserID, err = s.encryptTokenClaims(data.ID); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		var encryptedUserRoleID string
		if encryptedUserRoleID, err = s.encryptTokenClaims(data.RoleId); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		encodedEmail := encoding.Encode(data.Email)
		encodedUuidLogin := encoding.Encode(ctx.Auth.UuidLogin)

		tokenClaims := &modelToken.TokenClaims{
			ID:        encryptedUserID,
			RoleID:    encryptedUserRoleID,
			Email:     encodedEmail,
			UuidLogin: encodedUuidLogin,
			Exp:       time.Now().Add(time.Duration(24 * time.Hour)).Unix(),
		}
		authToken := modelToken.NewAuthToken(tokenClaims)
		token, err = authToken.Token()
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token": token,
	}, nil
}

func (s *service) SendEmailForgotPassword(ctx *abstraction.Context, payload *dto.AuthSendEmailForgotPasswordRequest) (map[string]interface{}, error) {
	var sendNotifTo []int = nil
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		data, err := s.UserRepository.FindByEmail(ctx, payload.Email)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if data == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "email not found")
		}

		eksternalToken := new(modelToken.AuthEksternalToken)
		eksternalToken.UserId = data.ID
		token, err := eksternalToken.GenerateTokenEksternal()
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		s.DbRedis.Set(context.Background(), *token, *token, 0)

		if err = gomail.SendMail(data.Email, "Forgot Password for ISS CleanCare", general.ParseTemplateEmailToHtml("./assets/html/email/notif_forgot_password.html", struct {
			NAME  string
			EMAIL string
			LINK  string
		}{
			NAME:  data.Name,
			EMAIL: data.Email,
			LINK:  constant.BASE_URL + "/auth/validation/reset-password/" + *token,
		})); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, v := range general.RemoveDuplicateArrayInt(sendNotifTo) {
		if err := ws.PublishNotificationWithoutTransaction(v, s.DB, ctx); err != nil {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
	}

	return map[string]interface{}{
		"message": "success send email forgot password!",
	}, nil
}

func (s *service) ValidationResetPassword(ctx *abstraction.Context, payload *dto.AuthValidationResetPasswordRequest) (string, error) {
	userData := new(model.UserEntityModel)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		_, err := s.DbRedis.Get(context.Background(), payload.Token).Result()
		if err == redis.Nil {
			return errors.New("your token is invalid")
		} else {
			s.DbRedis.Del(context.Background(), payload.Token)
		}

		data, err := modelToken.ValidateTokenEksternal(payload.Token)
		if err != nil {
			return errors.New("your token is invalid")
		}

		userData, err = s.UserRepository.FindById(ctx, data.UserId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		passwordString := general.GeneratePassword(8, 1, 1, 1, 1)
		password := []byte(passwordString)
		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.Password = string(hashedPassword)

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if err = gomail.SendMail(userData.Email, "Reset Password for ISS CleanCare", general.ParseTemplateEmailToHtml("./assets/html/email/notif_reset_password.html", struct {
			NAME      string
			RESETNAME string
			NUMBERID  string
			PASSWORD  string
			LINK      string
		}{
			NAME:      userData.Name,
			RESETNAME: "System",
			NUMBERID:  userData.Email,
			PASSWORD:  passwordString,
			LINK:      constant.BASE_URL,
		})); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return userData.Email, nil
}
