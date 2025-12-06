package controller

import (
	"errors"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"net/http"
)

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func internalError(c *gin.Context, err error) {
	c.Error(err) // отдаст в ErrorHandler()
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		badRequest(c, err)
		return uuid.Nil, false
	}
	return id, true
}

func getCtxInt64(c *gin.Context, key string) (int64, bool) {
	val, ok := c.Get(key)
	if !ok {
		badRequest(c, errors.New("missing context key: "+key))
		return 0, false
	}
	return val.(int64), true
}

func getCtxUUID(c *gin.Context, key string) (uuid.UUID, bool) {
	val, ok := c.Get(key)
	if !ok {
		badRequest(c, errors.New("missing context key: "+key))
		return uuid.Nil, false
	}
	id, err := uuid.Parse(val.(string))
	if err != nil {
		badRequest(c, err)
		return uuid.Nil, false
	}
	return id, true
}

type eventController struct {
	eventService            service.EventService
	eventParticipantService service.EventParticipantService
}

func (uc *eventController) getAllEvents(c *gin.Context) {
	vkId, ok := getCtxInt64(c, config.UserVkIdContextKey)
	if !ok {
		return
	}

	events, err := uc.eventService.GetAll(vkId)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, gin.H{"events": events})
}

func (uc *eventController) getEvent(c *gin.Context) {
	eventId, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	vkId, ok := getCtxInt64(c, config.UserVkIdContextKey)
	if !ok {
		return
	}

	event, err := uc.eventService.GetEvent(eventId, vkId)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, event)
}

func (uc *eventController) createEvent(c *gin.Context) {
	var updateEventDto dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		badRequest(c, err)
		return
	}

	event, err := uc.eventService.CreateEvent(updateEventDto)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, event)
}

func (uc *eventController) updateEvent(c *gin.Context) {
	eventId, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	var updateEventDto dto.CreateUpdateEventDto
	if err := c.ShouldBindJSON(&updateEventDto); err != nil {
		badRequest(c, err)
		return
	}

	event, err := uc.eventService.UpdateEvent(eventId, updateEventDto)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, event)
}

func (uc *eventController) getEventParticipants(c *gin.Context) {
	eventId, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	participants, err := uc.eventService.GetEventParticipants(eventId)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, gin.H{"participants": participants})
}

func (uc *eventController) getEventsAvailable(c *gin.Context) {
	roleId, ok := getCtxUUID(c, config.UserRoleContextKey)
	if !ok {
		return
	}

	events, err := uc.eventService.GetEventsByRole(roleId)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, gin.H{"events": events})
}

func (uc *eventController) getEventsArchive(c *gin.Context) {
	roleId, ok := getCtxUUID(c, config.UserRoleContextKey)
	if !ok {
		return
	}

	events, err := uc.eventService.GetArchiveEvents(roleId)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(200, gin.H{"events": events})
}

func (uc *eventController) participateToEvent(c *gin.Context) {
	eventId, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	vkId, ok := getCtxInt64(c, config.UserVkIdContextKey)
	if !ok {
		return
	}

	success, err := uc.eventParticipantService.ParticipantToEvent(eventId, vkId)
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(200, gin.H{"success": success})
}

func (uc *eventController) unParticipateToEvent(c *gin.Context) {
	eventId, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	vkId, ok := getCtxInt64(c, config.UserVkIdContextKey)
	if !ok {
		return
	}

	success, err := uc.eventParticipantService.UpParticipantToEvent(eventId, vkId)
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(200, gin.H{"success": success})
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

	r.GET("/events/all", ec.getAllEvents)
	r.GET("/events/:id", ec.getEvent)
	r.PUT("/events/:id", ec.updateEvent)
	r.GET("/events/:id/participants", ec.getEventParticipants)

	r.GET("/events/available", ec.getEventsAvailable)
	r.GET("/events/archive", ec.getEventsArchive)

	r.POST("/events/:id/participate", ec.participateToEvent)
	r.POST("/events/:id/unparticipate", ec.unParticipateToEvent)
}
