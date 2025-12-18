package middleware

import (
	"errors"
	"ne_noy/internal/config"
	"ne_noy/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AdminMiddleware(roleRepository repository.RoleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIdString, ok := c.Get(config.UserRoleContextKey)
		if !ok {
			c.JSON(400, gin.H{
				"error": errors.New("context does not contains role"),
			})
			c.Abort()
		}
		roleIdUuid, err := uuid.Parse(roleIdString.(string))
		role, err := roleRepository.GetById(roleIdUuid)
		if err != nil {
			c.Error(err)
			return
		}
		if role.Name != config.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "role name should be admin"})
			return
		}
		c.Next()
	}
}
