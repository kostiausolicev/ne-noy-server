package event

import (
	"ne_noy/internal/service/event"
	"ne_noy/internal/service/event/event_as_team"

	"github.com/gin-gonic/gin"
)

type teamController struct {
	eventService event.EventService
	teamService  event_as_team.EventTeamService
}

func (t *teamController) CreateTeam(c *gin.Context) {

}

func (t *teamController) JoinTeam(c *gin.Context) {

}

func (t *teamController) LeaveTeam(c *gin.Context) {

}

func (t *teamController) GetTeamsByEvent(c *gin.Context) {

}

func (t *teamController) GetTeam(c *gin.Context) {

}

func ConfigureTeamEventController(
	r *gin.RouterGroup,
	eventService event.EventService,
	teamService event_as_team.EventTeamService,
) {
	controller := &teamController{
		eventService: eventService,
		teamService:  teamService,
	}

	r.POST("/events/team", controller.CreateTeam)
	r.POST("/events/team/:id/joinTo/:teamId", controller.JoinTeam)
	r.POST("/events/team/:id/leaveFrom/:teamId", controller.LeaveTeam)
	r.GET("/events/team/:id", controller.GetTeamsByEvent)
}
