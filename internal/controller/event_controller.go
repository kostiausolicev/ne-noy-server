package controller

import (
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type eventController struct {
	eventService            service.EventService
	eventParticipantService service.EventParticipantService
}

func (uc *eventController) getEvent(c *gin.Context) {}

func (uc *eventController) getEventsAvailable(c *gin.Context) {

}

func (uc *eventController) getEventsAll(c *gin.Context) {
	roleIdStr, _ := c.Get("role_id")
	roleId, _ := uuid.Parse(roleIdStr.(string))
	events, err := uc.eventService.GetEventsByRole(roleId)
	if err != nil {
		return
	}
	c.JSON(200, events)
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
	router.GET("/events/all", ec.getEventsAll)
	router.POST("/events/:id/participate", ec.participateToEvent)
}
