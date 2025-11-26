package controller

import (
	"ne_noy/internal/config"
	"ne_noy/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type eventController struct {
	eventService            service.EventService
	eventParticipantService service.EventParticipantService
}

func (uc *eventController) getEvent(c *gin.Context) {
	eventId, _ := uuid.Parse(c.Param("id"))
	vkIdStr, _ := strconv.ParseInt(c.GetHeader(config.UserVkIdContextKey), 10, 64)

	event, err := uc.eventService.GetEvent(eventId, vkIdStr)
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

func (uc *eventController) participateToEvent(c *gin.Context) {

}

func (uc *eventController) unParticipateToEvent(c *gin.Context) {

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
	router.POST("/events/:id/participate", ec.participateToEvent)
}
