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

	routeTeamBase         = "/events/team"
	routeTeam             = "/events/team/:id"
	routeGetTeams         = "/events/team/:id/teams"
	routeTeamJoin         = "/events/team/:id/join/:teamId"
	routeTeamLeave        = "/events/team/:id/leave/:teamId"
	routeTeamByID         = "/events/team/:id/:teamId"
	routeTeamNotification = "/events/team/:id/:teamId/notification"

	routeTest                = "/events/test"
	routeTestByID            = "/events/test/:id"
	routeTestQuestion        = "/events/test/:id/q"
	routeTestQuestionByID    = "/events/test/:id/q/:qId"
	routeTestQuestionAnswers = "/events/test/:id/q/:qId/answers"

	routeTeamEventFull        = "/team-events"
	routeTeamEventsID         = "/team-events/:id"
	routeEventTestMyResults   = "/events/:eventId/test/my-results"
	routeEventTestUserResults = "/events/:eventId/test/user-results"
	routeEventTestReport      = "/events/:eventId/test/report"
)
