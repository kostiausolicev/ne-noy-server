package event

import (
	"net/http"

	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/dto/team_dto"
	appservice "ne_noy/internal/service"
	eventservice "ne_noy/internal/service/event"
	"ne_noy/internal/service/event/event_as_team"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type teamController struct {
	eventService eventservice.EventService
	teamService  event_as_team.EventTeamService
	userService  appservice.UserService
}

// CreateTeamEvent godoc
//
//	@Summary		Создать командное мероприятие
//	@Description	Создает запись мероприятия типа "команды" с настройками количества команд и вместимости.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param			request			body		team_dto.CreateTeamEventDto	true	"Данные командного мероприятия"
//	@Success		200				{object}	team_dto.TeamEventDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/team [post]
//	@Security		VkAuth
func (t *teamController) CreateTeamEvent(c *gin.Context) {
	createEventDto, ok := controller.BindJSON[team_dto.CreateTeamEventDto](c)
	if !ok {
		return
	}

	event, err := t.teamService.CreateTeamEvent(c.Request.Context(), createEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, event)
}

// UpdateTeamEvent godoc
//
//	@Summary		Обновить командное мероприятие
//	@Description	Частично обновляет запись мероприятия типа "команды".
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param			id				path		string						true	"UUID командного мероприятия"
//	@Param			request			body		team_dto.UpdateTeamEventDto	true	"Поля для обновления командного мероприятия"
//	@Success		200				{object}	team_dto.TeamEventDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404				{object}	dto.ErrorResponse	"Командное мероприятие не найдено"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id} [patch]
//	@Security		VkAuth
func (t *teamController) UpdateTeamEvent(c *gin.Context) {
	eventID, err := controller.ParseUUID(c, controller.ParamID)
	if err != nil {
		c.Error(err)
		return
	}

	updateEventDto, ok := controller.BindJSON[team_dto.UpdateTeamEventDto](c)
	if !ok {
		return
	}

	event, err := t.teamService.UpdateTeamEvent(c.Request.Context(), eventID, updateEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, event)
}

// DeleteTeamEvent godoc
//
//	@Summary		Удалить командное мероприятие
//	@Description	Удаляет запись мероприятия типа "команды" вместе с командами и участниками команд.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param			id				path	string	true	"UUID командного мероприятия"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure		401	{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404	{object}	dto.ErrorResponse	"Командное мероприятие не найдено"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id} [delete]
//	@Security		VkAuth
func (t *teamController) DeleteTeamEvent(c *gin.Context) {
	eventID, err := controller.ParseUUID(c, controller.ParamID)
	if err != nil {
		c.Error(err)
		return
	}

	if err = t.teamService.DeleteTeamEvent(c.Request.Context(), team_dto.DeleteTeamEventDto{ID: eventID}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// CreateTeam godoc
//
//	@Summary		Создать команду в командном мероприятии
//	@Description	Создает новую команду в мероприятии типа "команды"; капитаном назначается текущий авторизованный пользователь.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string							true	"Уникальный идентификатор запроса"
//	@Param			id				path		string							true	"UUID командного мероприятия"
//	@Param			request			body		team_dto.CreateTeamRequestDto	true	"Данные команды"
//	@Success		201				{object}	team_dto.TeamDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404				{object}	dto.ErrorResponse	"Мероприятие или пользователь не найдены"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id} [post]
//	@Security		VkAuth
func (t *teamController) CreateTeam(c *gin.Context) {
	eventID, err := controller.ParseUUID(c, controller.ParamID)
	if err != nil {
		c.Error(err)
		return
	}

	createTeamRequest, ok := controller.BindJSON[team_dto.CreateTeamRequestDto](c)
	if !ok {
		return
	}

	userID, err := t.currentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}
	createTeamDto := team_dto.CreateTeamDto{
		Name:      createTeamRequest.Name,
		CaptainID: userID,
	}

	team, err := t.teamService.CreateTeam(c.Request.Context(), eventID, createTeamDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, team)
}

// JoinTeam godoc
//
//	@Summary		Вступить в команду
//	@Description	Добавляет текущего авторизованного пользователя в выбранную команду командного мероприятия.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param			id				path	string	true	"UUID командного мероприятия"
//	@Param			teamId			path	string	true	"UUID команды"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse	"Некорректный UUID или команда заполнена"
//	@Failure		401	{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404	{object}	dto.ErrorResponse	"Команда или пользователь не найдены"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id}/joinTo/{teamId} [post]
//	@Security		VkAuth
func (t *teamController) JoinTeam(c *gin.Context) {
	if _, err := controller.ParseUUID(c, controller.ParamID); err != nil {
		c.Error(err)
		return
	}

	teamID, err := controller.ParseUUID(c, controller.ParamTeamID)
	if err != nil {
		c.Error(err)
		return
	}

	userID, err := t.currentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err = t.teamService.JoinTeam(c.Request.Context(), teamID, userID); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// LeaveTeam godoc
//
//	@Summary		Выйти из команды
//	@Description	Удаляет текущего авторизованного пользователя из выбранной команды.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param			id				path	string	true	"UUID командного мероприятия"
//	@Param			teamId			path	string	true	"UUID команды"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse	"Некорректный UUID или капитан пытается выйти из команды"
//	@Failure		401	{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404	{object}	dto.ErrorResponse	"Команда или пользователь не найдены"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id}/leaveFrom/{teamId} [post]
//	@Security		VkAuth
func (t *teamController) LeaveTeam(c *gin.Context) {
	if _, err := controller.ParseUUID(c, controller.ParamID); err != nil {
		c.Error(err)
		return
	}

	teamID, err := controller.ParseUUID(c, controller.ParamTeamID)
	if err != nil {
		c.Error(err)
		return
	}

	userID, err := t.currentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err = t.teamService.LeaveTeam(c.Request.Context(), teamID, userID); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// GetTeamsByEvent godoc
//
//	@Summary		Получить команды мероприятия
//	@Description	Возвращает список команд командного мероприятия с капитаном, первыми участниками и общим количеством участников.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			id				path		string	true	"UUID командного мероприятия"
//	@Success		200				{array}		team_dto.TeamDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404				{object}	dto.ErrorResponse	"Мероприятие не найдено"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id}/teams [get]
//	@Security		VkAuth
func (t *teamController) GetTeamsByEvent(c *gin.Context) {
	eventID, err := controller.ParseUUID(c, controller.ParamID)
	if err != nil {
		c.Error(err)
		return
	}

	teams, err := t.teamService.GetTeamsOnEvent(c.Request.Context(), eventID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, teams)
}

// GetTeamEvent godoc
//
//	@Summary		Получить командное мероприятие
//	@Description	Возвращает командное мероприятие
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			id				path		string	true	"UUID командного мероприятия"
//	@Success		200				{array}		team_dto.TeamEventDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404				{object}	dto.ErrorResponse	"Мероприятие не найдено"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id}/teams [get]
//	@Security		VkAuth
func (t *teamController) GetTeamEvent(c *gin.Context) {
	eventID, err := controller.ParseUUID(c, controller.ParamID)
	if err != nil {
		c.Error(err)
		return
	}

	team, err := t.teamService.GetTeamEvent(c.Request.Context(), eventID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, team)
}

// GetTeam godoc
//
//	@Summary		Получить команду
//	@Description	Возвращает одну команду с капитаном, первыми участниками и общим количеством участников.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			id				path		string	true	"UUID командного мероприятия"
//	@Param			teamId			path		string	true	"UUID команды"
//	@Success		200				{object}	team_dto.TeamDto
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404				{object}	dto.ErrorResponse	"Команда не найдена"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id}/{teamId} [get]
//	@Security		VkAuth
func (t *teamController) GetTeam(c *gin.Context) {
	if _, err := controller.ParseUUID(c, controller.ParamID); err != nil {
		c.Error(err)
		return
	}

	teamID, err := controller.ParseUUID(c, controller.ParamTeamID)
	if err != nil {
		c.Error(err)
		return
	}

	team, err := t.teamService.GetTeam(c.Request.Context(), teamID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, team)
}

// SendNotificationToTeam godoc
//
//	@Summary		Отправить уведомление команде
//	@Description	Отправляет текстовое VK-уведомление капитану и участникам выбранной команды.
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string								true	"Уникальный идентификатор запроса"
//	@Param			id				path	string								true	"UUID командного мероприятия"
//	@Param			teamId			path	string								true	"UUID команды"
//	@Param			request			body	team_dto.SendTeamNotificationDto	true	"Текст уведомления"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure		401	{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		404	{object}	dto.ErrorResponse	"Команда не найдена"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/events/team/{id}/{teamId}/notification [post]
//	@Security		VkAuth
func (t *teamController) SendNotificationToTeam(c *gin.Context) {
	if _, err := controller.ParseUUID(c, controller.ParamID); err != nil {
		c.Error(err)
		return
	}

	teamID, err := controller.ParseUUID(c, controller.ParamTeamID)
	if err != nil {
		c.Error(err)
		return
	}

	notificationDto, ok := controller.BindJSON[team_dto.SendTeamNotificationDto](c)
	if !ok {
		return
	}

	if err = t.teamService.SendNotificationToTeam(c.Request.Context(), teamID, notificationDto.Message); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

func (t *teamController) currentUserID(c *gin.Context) (uuid.UUID, error) {
	vkID, err := controller.GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		return uuid.Nil, err
	}

	// В токене лежит VK ID, а доменная логика команд работает с внутренним UUID пользователя.
	user, err := t.userService.GetUserByVkId(c.Request.Context(), vkID)
	if err != nil {
		return uuid.Nil, err
	}
	if user == nil || user.ID == nil {
		return uuid.Nil, controller.ParseError
	}

	return *user.ID, nil
}

func ConfigureTeamEventController(
	r *gin.RouterGroup,
	eventService eventservice.EventService,
	teamService event_as_team.EventTeamService,
	userService appservice.UserService,
) {
	controller := &teamController{
		eventService: eventService,
		teamService:  teamService,
		userService:  userService,
	}

	r.POST(routeTeamBase, controller.CreateTeamEvent)
	r.PATCH(routeTeam, controller.UpdateTeamEvent)
	r.DELETE(routeTeam, controller.DeleteTeamEvent)
	r.POST(routeTeam, controller.CreateTeam)
	r.POST(routeTeamJoin, controller.JoinTeam)
	r.POST(routeTeamLeave, controller.LeaveTeam)
	r.GET(routeTeam, controller.GetTeamsByEvent)
	r.GET(routeGetTeams, controller.GetTeamsByEvent)
	r.GET(routeTeamByID, controller.GetTeam)
	r.POST(routeTeamNotification, controller.SendNotificationToTeam)
}
