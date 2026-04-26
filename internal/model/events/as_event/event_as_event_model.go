package as_event

import "ne_noy/internal/model/events"

type AsEvent struct {
	events.EventProfile
	events.EventRelations
	VkPostID          *int64
	VkVoteID          *int64
	VkPollAnswerID    *int64
	Lat               *float64
	Lon               *float64
	Address           *string
	AdditionalAddress *string

	ParticipantsCount int
	EventParticipants []EventParticipant
}
