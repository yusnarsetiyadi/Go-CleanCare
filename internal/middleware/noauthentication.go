package middleware

import (
	"cleancare/pkg/constant"
	"cleancare/pkg/util/response"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func ResetPasswordIpCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		ip := c.RealIP()
		if ip == "::1" {
			ip = "localhost"
		}

		keys := fmt.Sprintf(constant.REDIS_REQUEST_RESET_PASSWORD_IP_KEYS, ip)
		value := dbRedis.Incr(c.Request().Context(), keys)
		if value.Err() != nil {
			return response.ErrorResponse(value.Err()).SendError(c)
		}

		if value.Val() > constant.REDIS_REQUEST_MAX_ATTEMPTS_RESET_PASSWORD {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("can't proceed request"), "too many attempts, please try again in 4 hours").SendError(c)
		}

		errRedis := dbRedis.Expire(c.Request().Context(), keys, constant.REDIS_REQUEST_IP_EXPIRE*time.Minute)
		if errRedis.Err() != nil {
			return response.ErrorResponse(errRedis.Err()).SendError(c)
		}

		return next(c)
	}
}

func VerifyNumberIpCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		ip := c.RealIP()
		if ip == "::1" {
			ip = "localhost"
		}

		keys := fmt.Sprintf(constant.REDIS_REQUEST_VERIFY_NUMBER_IP_KEYS, ip)
		value := dbRedis.Incr(c.Request().Context(), keys)
		if value.Err() != nil {
			return response.ErrorResponse(value.Err()).SendError(c)
		}

		if value.Val() > constant.REDIS_REQUEST_MAX_ATTEMPTS_VERIFY_NUMBER {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("can't proceed request"), "too many attempts, please try again in 4 hours").SendError(c)
		}

		errRedis := dbRedis.Expire(c.Request().Context(), keys, constant.REDIS_REQUEST_IP_EXPIRE*time.Minute)
		if errRedis.Err() != nil {
			return response.ErrorResponse(errRedis.Err()).SendError(c)
		}

		return next(c)
	}
}
