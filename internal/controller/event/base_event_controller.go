package event

import (
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/service/event"

	"github.com/gin-gonic/gin"
)

type baseEventController struct {
	eventService event.EventService
}

// getAllEvents godoc
//
//	@Summary	Получить список всех мероприятий
//	@Tags		base-events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventMiniDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events [get]
//	@Security	VkAuth
func (uc *baseEventController) getAllEvents(c *gin.Context) {
	vkId, err := controller.GetCtxInt64(c, config.UserVkIdContextKey)
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

// getEventsAvailable godoc
//
//	@Summary	Получить список всех доступных мероприятий
//	@Tags		base-events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventMiniDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/available [get]
//	@Security	VkAuth
func (uc *baseEventController) getEventsAvailable(c *gin.Context) {
	role, err := controller.GetCtxString(c, config.UserRoleContextKey)
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
//	@Tags		base-events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventMiniDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/archive [get]
//	@Security	VkAuth
func (uc *baseEventController) getEventsArchive(c *gin.Context) {
	role, err := controller.GetCtxString(c, config.UserRoleContextKey)
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

// publishEvent godoc
//
//	@Summary	Опубликовать мероприятие
//	@Tags		base-events
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param		id				path		string						true	"UUID мероприятия"
//	@Success	200
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{id}/publish [post]
//	@Security	VkAuth
func (uc *baseEventController) publishEvent(c *gin.Context) {
	eventId, err := controller.ParseUUID(c, controller.ParamID)
	if err != nil {
		c.Error(err)
		return
	}

	err = uc.eventService.PublishEvent(c.Request.Context(), eventId)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
}

func ConfigureBaseEventController(
	r *gin.RouterGroup,
	eventService event.EventService,
) {
	ec := &baseEventController{eventService: eventService}
	r.GET(routeEvents, ec.getAllEvents)
	r.GET(routeEventsAvailable, ec.getEventsAvailable)
	r.GET(routeEventsArchive, ec.getEventsArchive)
	r.POST(routeEventPublish, ec.publishEvent)
}
