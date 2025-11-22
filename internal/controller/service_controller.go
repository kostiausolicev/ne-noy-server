package controller

import (
	"ne_noy/internal/config"
	"ne_noy/internal/repository"
	"ne_noy/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type serviceController struct {
	jwtService     service.JWTService
	userRepository repository.UserRepository
}

func (sc *serviceController) generateToken(c *gin.Context) {
	vkId, err := strconv.ParseInt(c.Param("vkId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid vkId: " + err.Error()})
		return
	}
	user, err := sc.userRepository.GetByVkId(vkId)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal server error: " + err.Error()})
		return
	}
	if user == nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	claims := map[string]interface{}{
		config.UserVkIdContextKey: user.VkID,
		config.UserRoleContextKey: user.Role.ID,
	}
	token, err := sc.jwtService.GenerateToken(claims)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(200, gin.H{"token": token})
}

func (sc *serviceController) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func ConfigureServiceController(
	router *gin.RouterGroup,
	jwtService service.JWTService,
	userRepository repository.UserRepository) {
	sc := &serviceController{jwtService: jwtService, userRepository: userRepository}
	router.GET("/health", func(c *gin.Context) { sc.healthCheck(c) })
	router.POST("/generateToken/:vkId", func(c *gin.Context) { sc.generateToken(c) })
}
