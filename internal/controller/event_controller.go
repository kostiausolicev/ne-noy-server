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

func (uc *eventController) GetEvent(c *gin.Context) {}

func (uc *eventController) GetEventsAvailable(c *gin.Context) {

}

func (uc *eventController) GetEventsAll(c *gin.Context) {
	roleIdStr, _ := c.Get("role_id")
	roleId, _ := uuid.Parse(roleIdStr.(string))
	events, err := uc.eventService.GetEventsByRole(roleId)
	if err != nil {
		return
	}
	c.JSON(200, events)
}

func (uc *eventController) ParticipateToEvent(c *gin.Context) {

}

func (uc *eventController) UnParticipateToEvent(c *gin.Context) {

}

func ConfigureEventController(
	router *gin.Engine,
	eventService service.EventService,
	eventParticipantService service.EventParticipantService) {
	ec := &eventController{
		eventService:            eventService,
		eventParticipantService: eventParticipantService,
	}
	router.GET("/events/:id", ec.GetEvent)
	router.GET("/events/available", ec.GetEventsAvailable)
	router.GET("/events/all", ec.GetEventsAll)
	router.POST("/events/:id/participate", ec.ParticipateToEvent)
}
