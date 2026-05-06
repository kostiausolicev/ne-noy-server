package event

import (
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/event_dto"
	"ne_noy/internal/service/event"
	"ne_noy/internal/service/event/event_as_event"

	"github.com/gin-gonic/gin"
)

type eventController struct {
	eventBaseService        event.EventService
	eventService            event_as_event.EventAsEventService
	eventParticipantService event_as_event.EventParticipantService
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
//	@Router		/v1/events/event/{id} [get]
//	@Security	VkAuth
func (uc *eventController) getEvent(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, "id")
	if err != nil {
		return
	}

	event, err := uc.eventService.GetEventById(c.Request.Context(), eventId)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, event)
}

// publishEvent godoc
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
//	@Router		/v1/events/event/{id} [patch]
//	@Security	VkAuth
func (uc *eventController) updateEvent(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	var updateEventDto event_dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		c.Error(err)
		return
	}

	updateEvent, err := uc.eventService.UpdateEvent(c.Request.Context(), eventId, updateEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, updateEvent)
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
//	@Router		/v1/events/event [post]
//	@Security	VkAuth
func (uc *eventController) createEvent(c *gin.Context) {
	var updateEventDto event_dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		c.Error(err)
		return
	}

	createEvent, err := uc.eventService.CreateEvent(c.Request.Context(), updateEventDto)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, createEvent)
}

// publishEvent godoc
//
//	@Summary	Опубликовать мероприятие
//	@Tags		events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		id				path		string	true	"UUID мероприятия для запроса"
//	@Success	200
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/publish [post]
//	@Security	VkAuth
func (uc *eventController) publishEvent(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	var updateEventDto event_dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		c.Error(err)
		return
	}

	err = uc.eventBaseService.PublishEvent(c.Request.Context(), eventId)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
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
//	@Router		/v1/events/event/{id}/participants [get]
//	@Security	VkAuth
func (uc *eventController) getEventParticipants(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, "id")
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
//	@Router		/v1/events/event/{id}/participate [post]
//	@Security	VkAuth
func (uc *eventController) participateToEvent(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	vkId, err := controller.GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	_, err = uc.eventParticipantService.ParticipantToEvent(c.Request.Context(), eventId, vkId, "app")
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
//	@Router		/v1/events/event/{id}/unparticipate [post]
//	@Security	VkAuth
func (uc *eventController) unParticipateToEvent(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	vkId, err := controller.GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	_, err = uc.eventParticipantService.UnParticipantToEvent(c.Request.Context(), eventId, vkId)
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
//	@Router		/v1/events/event/{id}/participate/check [patch]
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
	eventService event.EventService,
	eventParticipantService event_as_event.EventParticipantService,
) {
	ec := &eventController{
		eventBaseService:        eventService,
		eventParticipantService: eventParticipantService,
	}

	// Запросы к типу "мероприятие"
	r.POST("/events/event", ec.createEvent)
	r.GET("/events/event/:id", ec.getEvent)
	r.PATCH("/events/event/:id", ec.updateEvent)
	r.GET("/events/event/:id/participants", ec.getEventParticipants)
	r.POST("/events/event/:id/participate", ec.participateToEvent)
	r.POST("/events/event/:id/unparticipate", ec.unParticipateToEvent)
	r.PATCH("/events/event/:id/participate/check", ec.checkParticipate)

	// Запросы к типу "команда"
}
