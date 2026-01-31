package middleware

import (
	"errors"
	"ne_noy/internal/controller"
	"ne_noy/internal/dto"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			requestId := c.GetHeader("X-Request-Id")
			if errors.Is(err, controller.ParseError) {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error(), Timestamp: time.Now(), RequestId: requestId})
			} else if errors.Is(err, controller.ForbiddenError) {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error(), Timestamp: time.Now(), RequestId: requestId})
			} else if errors.Is(err, controller.AuthorizationError) {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error(), Timestamp: time.Now(), RequestId: requestId})
			} else if errors.Is(err, controller.InvalidAuthTokenError) {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error(), Timestamp: time.Now(), RequestId: requestId})
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error(), Timestamp: time.Now(), RequestId: requestId})
			} else {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error(), Timestamp: time.Now(), RequestId: requestId})
			}
		}
	}
}
