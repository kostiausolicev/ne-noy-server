package controller

import (
	"ne_noy/internal/repository"

	"github.com/gin-gonic/gin"
)

type serviceController struct {
	userRepository repository.UserRepository
}

func (sc *serviceController) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func ConfigureServiceController(
	router *gin.RouterGroup,
	userRepository repository.UserRepository) {
	sc := &serviceController{userRepository: userRepository}
	router.GET("/health", func(c *gin.Context) { sc.healthCheck(c) })
}
