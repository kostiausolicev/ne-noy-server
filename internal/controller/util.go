package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ParamID         = "id"
	ParamTeamID     = "teamId"
	ParamQuestionID = "qId"
	ParamPermission = "permission"
	ParamRoleID     = "roleId"

	QueryFIO      = "fio"
	QueryPlatform = "platform"
	QueryValue    = "value"

	HeaderAuthorization = "authorization"
	HeaderRequestID     = "X-Request-Id"
)

var (
	ParseError            = errors.New("parse error")
	ForbiddenError        = errors.New("forbidden")
	AuthorizationError    = errors.New("authorization error")
	InvalidAuthTokenError = errors.New("invalid auth token")
)

func ParseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	return id, err
}

func ParseInt64Param(c *gin.Context, param string) (int64, error) {
	value, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseBoolQuery(c *gin.Context, query string) (bool, error) {
	value, err := strconv.ParseBool(c.Query(query))
	if err != nil {
		return false, err
	}
	return value, nil
}

func BindJSON[T any](c *gin.Context) (T, bool) {
	var payload T
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(err)
		return payload, false
	}
	return payload, true
}

func GetCtxInt64(c *gin.Context, key string) (int64, error) {
	val, ok := c.Get(key)
	if !ok {
		return 0, ParseError
	}
	typedVal, ok := val.(int64)
	if !ok {
		return 0, ParseError
	}
	return typedVal, nil
}

func GetCtxString(c *gin.Context, key string) (string, error) {
	val, ok := c.Get(key)
	if !ok {
		return "", ParseError
	}
	typedVal, ok := val.(string)
	if !ok {
		return "", ParseError
	}
	return typedVal, nil
}
