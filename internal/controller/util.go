package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ParseError = errors.New("parse error")
var ForbiddenError = errors.New("forbidden")
var AuthorizationError = errors.New("authorization error")
var InvalidAuthTokenError = errors.New("invalid auth token")

func ParseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	return id, err
}

func GetCtxInt64(c *gin.Context, key string) (int64, error) {
	val, ok := c.Get(key)
	if !ok {
		return 0, ParseError
	}
	return val.(int64), nil
}

func GetCtxUUID(c *gin.Context, key string) (uuid.UUID, error) {
	val, ok := c.Get(key)
	if !ok {
		return uuid.Nil, ParseError
	}
	id, err := uuid.Parse(val.(string))
	if err != nil {
		return uuid.Nil, ParseError
	}
	return id, nil
}
