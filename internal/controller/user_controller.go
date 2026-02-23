package controller

import (
	"net/http"
	"strconv"

	"ne_noy/internal/dto"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userController struct {
	service service.UserService
}

// createUser godoc
//
//	@Summary		Создать нового пользователя
//	@Description	Создаёт пользователя на основе переданных данных
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string		true	"Уникальный идентификатор запроса для трассировки"
//	@Param			user			body		dto.CreateUserDto	true	"Данные нового пользователя"
//	@Success		201				{object}	dto.UserDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные входные данные"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		409				{object}	dto.ErrorResponse	"Пользователь уже существует"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/users [post]
//	@Security		VkAuth
func (uc *userController) createUser(c *gin.Context) {
	var user dto.CreateUserDto
	if err := c.ShouldBindJSON(&user); err != nil {
		c.Error(err)
		return
	}
	createUser, err := uc.service.CreateUser(c.Request.Context(), user)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, createUser)
}

// getAll godoc
//
//	@Summary		Получить список пользователей
//	@Description	Возвращает всех пользователей с возможной фильтрацией по ФИО
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			fio				query		string	false	"Фильтр по ФИО (частичное совпадение)"
//	@Success		200				{array}		dto.UserDto
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/users [get]
//	@Security		VkAuth
func (uc *userController) getAll(c *gin.Context) {
	fio := c.Query("fio")
	users, err := uc.service.GetAllUsers(c.Request.Context(), fio)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// getByVkId godoc
//
//	@Summary		Получить пользователя по VK ID
//	@Description	Находит пользователя по его VK ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			id				path		string	true	"VK ID пользователя (числовой идентификатор, например: 12345678)"
//	@Success		200				{object}	dto.UserDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректный формат VK ID"
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		404				{object}	dto.ErrorResponse	"Пользователь не найден"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/users/vk/{id} [get]
//	@Security		VkAuth
func (uc *userController) getByVkId(c *gin.Context) {
	vkIdStr := c.Param("id")
	vkId, err := strconv.ParseInt(vkIdStr, 10, 64)
	if err != nil {
		c.Error(err)
		return
	}
	user, err := uc.service.GetUserByVkId(c.Request.Context(), vkId)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, user)
}

// updateUser godoc
//
//	@Summary		Обновить одно разрешение (permission) пользователя
//	@Description	Изменяет значение конкретного флага-разрешения (true/false)
//	@Tags			users admin
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param			id				path	string	true	"VK ID пользователя"
//	@Param			permission		path	string	true	"Название разрешения (например: is_admin, can_moderate и т.д.)"
//	@Param			value			query	boolean	true	"Новое значение разрешения (true/false)"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse	"Некорректные параметры"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse	"Нет прав на изменение"
//	@Failure		404	{object}	dto.ErrorResponse	"Пользователь не найден"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/users/vk/{id}/{permission} [patch]
//	@Security		VkAuth
func (uc *userController) updateUser(c *gin.Context) {
	vkIdStr := c.Param("id")
	vkId, err := strconv.ParseInt(vkIdStr, 10, 64)
	if err != nil {
		c.Error(err)
		return
	}
	permission := c.Param("permission")
	valueStr := c.Query("value")
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		c.Error(err)
		return
	}
	err = uc.service.UpdatePermissions(c.Request.Context(), permission, vkId, value)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// updateUserRole godoc
//
//	@Summary		Обновить роль пользователя
//	@Description	Назначает пользователю новую роль по её UUID
//	@Tags			users admin
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param			id				path	string	true	"VK ID пользователя"
//	@Param			roleId			path	string	true	"UUID новой роли (формат: 550e8400-e29b-41d4-a716-446655440000)"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse	"Некорректный UUID роли или VK ID"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse	"Нет прав"
//	@Failure		404	{object}	dto.ErrorResponse	"Пользователь или роль не найдены"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/users/vk/{id}/role/{roleId} [patch]
//	@Security		VkAuth
func (uc *userController) updateUserRole(c *gin.Context) {
	vkIdStr := c.Param("id")
	vkId, err := strconv.ParseInt(vkIdStr, 10, 64)
	if err != nil {
		c.Error(err)
		return
	}
	roleIdStr := c.Param("roleId")
	roleId, err := uuid.Parse(roleIdStr)
	if err != nil {
		c.Error(err)
		return
	}
	err = uc.service.UpdateRole(c.Request.Context(), vkId, roleId)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

func ConfigureUserController(router *gin.RouterGroup, service service.UserService) {
	uc := &userController{service: service}
	router.POST("/users", uc.createUser)
	router.GET("/users", uc.getAll)
	router.GET("/users/vk/:id", uc.getByVkId)
	router.PATCH("/users/vk/:id/:permission", uc.updateUser)
}

func ConfigureAdminUserController(router *gin.RouterGroup, service service.UserService) {
	uc := &userController{service: service}
	router.PATCH("/users/vk/:id/role/:roleId", uc.updateUserRole)
}
