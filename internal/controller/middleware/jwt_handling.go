package middleware

import (
	"ne_noy/internal/config"

	"github.com/gin-gonic/gin"
)

// TODO заменить jwt на проверку vk токена
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetHeader(config.UserRoleContextKey)
		vkId := c.GetHeader(config.UserVkIdContextKey)

		c.Set(config.UserRoleContextKey, role)
		c.Set(config.UserVkIdContextKey, vkId)

		c.Next()
	}
}
