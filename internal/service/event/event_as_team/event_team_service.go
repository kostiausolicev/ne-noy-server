package event_as_team

type EventTeamService interface {
	GetTeamsOnEvent()
	CreateTeam()
	JoinTeam()
	LeaveTeam()
}
