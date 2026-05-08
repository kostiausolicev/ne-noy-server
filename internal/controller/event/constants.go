package event

const (
	participantSourceApp = "app"

	routeEvents          = "/events"
	routeEventsAvailable = "/events/available"
	routeEventsArchive   = "/events/archive"
	routeEventPublish    = "/events/:id/publish"

	routeEvent                 = "/events/event"
	routeEventByID             = "/events/event/:id"
	routeEventParticipants     = "/events/event/:id/participants"
	routeEventParticipate      = "/events/event/:id/participate"
	routeEventUnparticipate    = "/events/event/:id/unparticipate"
	routeEventCheckParticipate = "/events/event/:id/participate/check"

	routeTeam             = "/events/team/:id"
	routeTeamJoin         = "/events/team/:id/joinTo/:teamId"
	routeTeamLeave        = "/events/team/:id/leaveFrom/:teamId"
	routeTeamByID         = "/events/team/:id/:teamId"
	routeTeamNotification = "/events/team/:id/:teamId/notification"

	routeTest                = "/events/test"
	routeTestByID            = "/events/test/:id"
	routeTestQuestion        = "/events/test/:id/q"
	routeTestQuestionByID    = "/events/test/:id/q/:qId"
	routeTestQuestionAnswers = "/events/test/:id/q/:qId/answers"
)
