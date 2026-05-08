package middleware

import (
	"ne_noy/internal/config"
	"ne_noy/internal/controller"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := controller.GetCtxString(c, config.UserRoleContextKey)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if role != config.RoleAdmin {
			c.Error(controller.ForbiddenError)
			c.Abort()
			return
		}
		c.Next()
	}
}
