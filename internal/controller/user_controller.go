package controller

import (
	"net/http"

	"ne_noy/internal/dto"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
)

type userController struct {
	service service.UserService
}

const (
	routeUsers          = "/users"
	routeUsersByLinks   = "/users/byLinks"
	routeUserByVkID     = "/users/vk/:id"
	routeUserPermission = "/users/vk/:id/:permission"
	routeUserRole       = "/users/vk/:id/role/:roleId"
)

// createUser godoc
//
//	@Summary		Создать нового пользователя
//	@Description	Создаёт пользователя на основе переданных данных
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string				true	"Уникальный идентификатор запроса для трассировки"
//	@Param			user			body		dto.CreateUserDto	true	"Данные нового пользователя"
//	@Success		201				{object}	dto.UserDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные входные данные"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		409				{object}	dto.ErrorResponse	"Пользователь уже существует"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/users [post]
//	@Security		VkAuth
func (uc *userController) createUser(c *gin.Context) {
	user, ok := BindJSON[dto.CreateUserDto](c)
	if !ok {
		return
	}
	createUser, err := uc.service.CreateUser(c.Request.Context(), user)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, createUser)
}

// createUserByLinks godoc
//
//	@Summary		Создать пользователей из списка ссылок
//	@Description	Создаёт пользоватей из списка ссылок
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string						true	"Уникальный идентификатор запроса для трассировки"
//	@Param			user			body		dto.CreateUserByLinksDto	true	"Ссылки на новых пользователей"
//	@Success		201				{array}		dto.UserDto
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/users/byLinks [post]
//	@Security		VkAuth
func (uc *userController) createUserByLinks(c *gin.Context) {
	users, ok := BindJSON[dto.CreateUserByLinksDto](c)
	if !ok {
		return
	}
	createUsers, err := uc.service.CreateUserByLinks(c.Request.Context(), users.Links)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, createUsers)
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
	fio := c.Query(QueryFIO)
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
	vkId, err := ParseInt64Param(c, ParamID)
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
	vkId, err := ParseInt64Param(c, ParamID)
	if err != nil {
		c.Error(err)
		return
	}
	permission := c.Param(ParamPermission)
	value, err := ParseBoolQuery(c, QueryValue)
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
	vkId, err := ParseInt64Param(c, ParamID)
	if err != nil {
		c.Error(err)
		return
	}
	roleId, err := ParseUUID(c, ParamRoleID)
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
	router.POST(routeUsers, uc.createUser)
	router.POST(routeUsersByLinks, uc.createUserByLinks)
	router.GET(routeUsers, uc.getAll)
	router.GET(routeUserByVkID, uc.getByVkId)
	router.PATCH(routeUserPermission, uc.updateUser)
}

func ConfigureAdminUserController(router *gin.RouterGroup, service service.UserService) {
	uc := &userController{service: service}
	router.PATCH(routeUserRole, uc.updateUserRole)
}
