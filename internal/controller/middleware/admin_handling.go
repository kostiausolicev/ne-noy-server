package middleware

import (
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/repository"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware(roleRepository repository.RoleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIdUuid, err := controller.GetCtxUUID(c, config.UserRoleContextKey)
		role, err := roleRepository.GetById(roleIdUuid)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if role.Name != config.RoleAdmin {
			c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}
