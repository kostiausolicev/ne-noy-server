package controller

import (
	"ne_noy/internal/config"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type eventController struct {
	eventService            service.EventService
	eventParticipantService service.EventParticipantService
}

func (uc *eventController) getEvent(c *gin.Context) {
	eventId, _ := uuid.Parse(c.Param("id"))
	vkId, _ := c.Get(config.UserVkIdContextKey)

	event, err := uc.eventService.GetEvent(eventId, vkId.(int64))
	if err != nil {
		return
	}
	c.JSON(200, event)
}

func (uc *eventController) getEventsAvailable(c *gin.Context) {
	roleIdStr, _ := c.Get(config.UserRoleContextKey)
	roleId, _ := uuid.Parse(roleIdStr.(string))
	events, err := uc.eventService.GetEventsByRole(roleId)
	if err != nil {
		return
	}
	c.JSON(200, gin.H{
		"events": events,
	})
}

func (uc *eventController) getEventsArchive(c *gin.Context) {
	roleIdStr, _ := c.Get(config.UserRoleContextKey)
	roleId, _ := uuid.Parse(roleIdStr.(string))
	events, err := uc.eventService.GetArchiveEvents(roleId)
	if err != nil {
		return
	}
	c.JSON(200, gin.H{
		"events": events,
	})
}

func (uc *eventController) participateToEvent(c *gin.Context) {
	eventId, _ := uuid.Parse(c.Param("id"))
	vkId, _ := c.Get(config.UserVkIdContextKey)
	success, err := uc.eventParticipantService.ParticipantToEvent(eventId, vkId.(int64))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": success,
	})
}

func (uc *eventController) unParticipateToEvent(c *gin.Context) {
	eventId, _ := uuid.Parse(c.Param("id"))
	vkId, _ := c.Get(config.UserVkIdContextKey)
	success, err := uc.eventParticipantService.UpParticipantToEvent(eventId, vkId.(int64))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": success,
	})
}

func ConfigureEventController(
	router *gin.RouterGroup,
	eventService service.EventService,
	eventParticipantService service.EventParticipantService) {
	ec := &eventController{
		eventService:            eventService,
		eventParticipantService: eventParticipantService,
	}
	router.GET("/events/:id", ec.getEvent)
	router.GET("/events/available", ec.getEventsAvailable)
	router.GET("/events/archive", ec.getEventsArchive)
	router.POST("/events/:id/participate", ec.participateToEvent)
	router.POST("/events/:id/unparticipate", ec.unParticipateToEvent)
}
