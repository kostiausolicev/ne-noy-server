package controller

import (
	"ne_noy/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type userController struct {
	service service.UserService
}

func (uc *userController) CreateUser(c *gin.Context) {
	// ...
}

func (uc *userController) GetAll(c *gin.Context) {
	vkId := c.Query("vk_id")
	if vkId != "" {
		vkIdLong, err := strconv.ParseInt(vkId, 10, 64)
		if err != nil {
			c.Error(err)
			return
		}
		user, err := uc.service.GetUserByVkId(vkIdLong)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	} else {

	}
}

func (uc *userController) GetById(c *gin.Context) {
	// ...
}

func (uc *userController) UpdateUser(c *gin.Context) {
	// ...
}

func (uc *userController) DeleteUser(c *gin.Context) {
	// ...
}

func ConfigureUserController(router *gin.Engine, service service.UserService) {
	uc := &userController{service: service}
	router.POST("/users", uc.CreateUser)
	router.GET("/users", uc.GetAll)
	router.GET("/users/:id", uc.GetById)
	router.PUT("/users/:id", uc.UpdateUser)
	router.DELETE("/users/:id", uc.DeleteUser)
}
