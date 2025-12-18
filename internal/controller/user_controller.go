package controller

import (
	"ne_noy/internal/dto"
	"ne_noy/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	fio := c.Query("fio")
	users, err := uc.service.GetAllUsers(fio)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (uc *userController) GetByVkId(c *gin.Context) {
	vkId := c.Param("id")
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
}

func (uc *userController) GetById(c *gin.Context) {
	// ...
}

func (uc *userController) UpdateUser(c *gin.Context) {
	vkId := c.Param("id")
	vkIdLong, err := strconv.ParseInt(vkId, 10, 64)
	if err != nil {
		c.Error(err)
		return
	}
	permission := c.Param("permission")
	value, err := strconv.ParseBool(c.Query("value"))
	if err != nil {
		c.Error(err)
		return
	}
	err = uc.service.UpdatePermissions(permission, vkIdLong, value)
}

func (uc *userController) UpdateUserRole(c *gin.Context) {
	vkId := c.Param("id")
	vkIdLong, err := strconv.ParseInt(vkId, 10, 64)
	if err != nil {
		c.Error(err)
		return
	}
	roleIdString := c.Param("roleId")
	roleIdUuid, err := uuid.Parse(roleIdString)
	err = uc.service.UpdateRole(vkIdLong, roleIdUuid)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (uc *userController) DeleteUser(c *gin.Context) {
	// ...
}

func ConfigureUserController(router *gin.RouterGroup, service service.UserService) {
	uc := &userController{service: service}
	router.POST("/users", uc.CreateUser)
	router.GET("/users", uc.GetAll)
	router.GET("/users/:id", uc.GetById)
	router.GET("/users/vk/:id", uc.GetByVkId)
	router.PUT("/users/vk/:id/:permission", uc.UpdateUser)
	router.DELETE("/users/vk/:id", uc.DeleteUser)
}

func ConfigureAdminUserController(router *gin.RouterGroup, service service.UserService) {
	uc := &userController{service: service}
	router.PUT("/users/vk/:id/role/:roleId", uc.UpdateUserRole)
}
