package controller

import (
	"ne_noy/internal/dto"
	"ne_noy/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type userController struct {
	service service.UserService
}

func (uc *userController) CreateUser(c *gin.Context) {
	var user dto.UserDto

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createUser, err := uc.service.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createUser)
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
		c.JSON(http.StatusOK, user)
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

func ConfigureUserController(router *gin.RouterGroup, service service.UserService) {
	uc := &userController{service: service}
	router.POST("/users", uc.CreateUser)
	router.GET("/users", uc.GetAll)
	router.GET("/users/:id", uc.GetById)
	router.PUT("/users/:id", uc.UpdateUser)
	router.DELETE("/users/:id", uc.DeleteUser)
}
