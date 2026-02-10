package controller

import (
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
)

type eventController struct {
	eventService            service.EventService
	eventParticipantService service.EventParticipantService
}

// getAllEvents godoc
//
//	@Summary	Получить список всех мероприятий
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventMiniDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events [get]
//	@Security	VkAuth
func (uc *eventController) getAllEvents(c *gin.Context) {
	vkId, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	events, err := uc.eventService.GetAll(c.Request.Context(), vkId)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, events)
}

// getEvent godoc
//
//	@Summary	Получить одно мероприятие по ID
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса для трассировки"
//	@Param		id				path		string	true	"UUID мероприятия (формат: 550e8400-e29b-41d4-a716-446655440000)"
//	@Success	200				{object}	dto.EventDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Мероприятие не найдено"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{id} [get]
//	@Security	VkAuth
func (uc *eventController) getEvent(c *gin.Context) {
	eventId, err := ParseUUID(c, "id")
	if err != nil {
		return
	}

	vkId, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		return
	}

	event, err := uc.eventService.GetEvent(c.Request.Context(), eventId, vkId)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, event)
}

// createEvent godoc
//
//	@Summary	Создать мероприятие
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string						true	"X-Request-Id"
//	@Param		request			body		dto.CreateUpdateEventDto	true	"дто для создания мероприятия"
//	@Success	200				{object}	dto.EventDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events [post]
//	@Security	VkAuth
func (uc *eventController) createEvent(c *gin.Context) {
	var updateEventDto dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		c.Error(err)
		return
	}

	event, err := uc.eventService.CreateEvent(c.Request.Context(), updateEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, event)
}

// publishEvent godoc
//
//	@Summary	Опубликовать мероприятие
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		id				path		string	true	"UUID мероприятия для запроса"
//	@Success	200				{object}	dto.EventDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/publish [post]
//	@Security	VkAuth
func (uc *eventController) publishEvent(c *gin.Context) {
	eventId, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	var updateEventDto dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		c.Error(err)
		return
	}

	event, err := uc.eventService.UpdateEvent(c.Request.Context(), eventId, updateEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, event)
}

// updateEvent godoc
//
//	@Summary	Обновить мероприятие
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param		id				path		string						true	"UUID мероприятия для обновления"
//	@Param		request			body		dto.CreateUpdateEventDto	true	"Данные для обновления мероприятия"
//	@Success	200				{object}	dto.EventDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{id} [patch]
//	@Security	VkAuth
func (uc *eventController) updateEvent(c *gin.Context) {
	eventId, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	var updateEventDto dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		c.Error(err)
		return
	}

	event, err := uc.eventService.UpdateEvent(c.Request.Context(), eventId, updateEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, event)
}

// getEventParticipants godoc
//
//	@Summary	Получить список участников мероприятия
//	@Tags		events users
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id для логирования"
//	@Param		id				path		string	true	"UUID мероприятия"
//	@Success	200				{array}		dto.EventParticipantDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/participants [get]
//	@Security	VkAuth
func (uc *eventController) getEventParticipants(c *gin.Context) {
	eventId, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	participants, err := uc.eventService.GetEventParticipants(c.Request.Context(), eventId)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, participants)
}

// getEventsAvailable godoc
//
//	@Summary	Получить список всех доступных мероприятий
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventMiniDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/available [get]
//	@Security	VkAuth
func (uc *eventController) getEventsAvailable(c *gin.Context) {
	role, err := GetCtxString(c, config.UserRoleContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	events, err := uc.eventService.GetEventsByRole(c.Request.Context(), role)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, events)
}

// getEventsArchive godoc
//
//	@Summary	Получить список всех архивных мероприятий
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventMiniDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/archive [get]
//	@Security	VkAuth
func (uc *eventController) getEventsArchive(c *gin.Context) {
	role, err := GetCtxString(c, config.UserRoleContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	events, err := uc.eventService.GetArchiveEvents(c.Request.Context(), role)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, events)
}

// participateToEvent godoc
//
//	@Summary	Участвовать в мероприятии
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header	string	true	"X-Request-Id"
//	@Param		id				path	string	true	"UUID мероприятия, в котором участвовать"
//	@Success	200
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Failure	500	{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/participate [post]
//	@Security	VkAuth
func (uc *eventController) participateToEvent(c *gin.Context) {
	eventId, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	vkId, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	_, err = uc.eventParticipantService.ParticipantToEvent(c.Request.Context(), eventId, vkId)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
}

// unParticipateToEvent godoc
//
//	@Summary	Отказаться от участия в мероприятии
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header	string	true	"X-Request-Id"
//	@Param		id				path	string	true	"UUID мероприятия"
//	@Success	200
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Failure	500	{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/unparticipate [post]
//	@Security	VkAuth
func (uc *eventController) unParticipateToEvent(c *gin.Context) {
	eventId, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	vkId, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	_, err = uc.eventParticipantService.UpParticipantToEvent(c.Request.Context(), eventId, vkId)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
}

// checkParticipate godoc
//
//	@Summary	Подтвердить/проверить участие (отметка)
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header	string						true	"X-Request-Id"
//	@Param		id				path	string						true	"UUID мероприятия"
//	@Param		request			body	dto.CheckEventParticipant	true	"Данные для проверки/отметки участия"
//	@Success	200
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	500	{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/participate/check [patch]
//	@Security	VkAuth
func (uc *eventController) checkParticipate(c *gin.Context) {
	var checkEventDto dto.CheckEventParticipant
	if err := c.ShouldBindJSON(&checkEventDto); err != nil {
		c.Error(err)
		return
	}
	err := uc.eventParticipantService.CheckParticipant(c.Request.Context(), checkEventDto)
	if err != nil {
		c.Error(err)
	}
	c.Status(200)
}

func ConfigureEventController(
	r *gin.RouterGroup,
	eventService service.EventService,
	eventParticipantService service.EventParticipantService,
) {
	ec := &eventController{
		eventService:            eventService,
		eventParticipantService: eventParticipantService,
	}

	r.POST("/events", ec.createEvent)

	r.GET("/events", ec.getAllEvents)
	r.GET("/events/:id", ec.getEvent)
	r.POST("/events/:id/publish", ec.updateEvent)
	r.PATCH("/events/:id", ec.updateEvent)
	r.GET("/events/:id/participants", ec.getEventParticipants)

	r.GET("/events/available", ec.getEventsAvailable)
	r.GET("/events/archive", ec.getEventsArchive)

	r.POST("/events/:id/participate", ec.participateToEvent)
	r.POST("/events/:id/unparticipate", ec.unParticipateToEvent)

	r.PATCH("/events/:id/participate/check", ec.checkParticipate)
}
