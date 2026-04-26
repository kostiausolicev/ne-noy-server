package repository

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"

	"github.com/google/uuid"
)

// EventRepository describes persistence operations for events and as_event profiles.
type EventRepository interface {
	// GetAll returns all events visible in the system.
	GetAll(ctx context.Context) ([]*events.EventView, error)

	// GetAllByOrg returns all events where the specified user is an organizer.
	GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*events.EventView, error)

	// GetAllByRole returns active events available for the specified role.
	GetAllByRole(ctx context.Context, role string) ([]*events.EventView, error)

	// GetAllArchive returns archived active events available for the specified role.
	GetAllArchive(ctx context.Context, role string) ([]*events.EventView, error)

	// GetEventLocationData returns location data for an as_event by its identifier.
	GetEventLocationData(ctx context.Context, id uuid.UUID) (*events.EventView, error)

	// GetEventTypeById returns the profile type of an as_event.
	GetEventTypeById(ctx context.Context, id uuid.UUID) (string, error)

	// GetUserParticipationInEvent checks whether a VK user participates in the as_event.
	GetUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userVkId int64) (bool, error)

	// GetEventOrgs returns organizers of the specified as_event.
	GetEventOrgs(ctx context.Context, eventId uuid.UUID) ([]model.User, error)

	// GetByVkPollId returns an as_event linked to the specified VK poll identifier.
	GetByVkPollId(ctx context.Context, pollId int64) (*events.EventView, error)

	// GetEventById returns a classic as_event profile from event_as_events.
	GetEventById(ctx context.Context, id uuid.UUID) (*model.EventAsEvent, error)

	// GetActivityById returns an activity as_event profile from event_as_activities.
	GetActivityById(ctx context.Context, id uuid.UUID) (*model.EventAsActivity, error)

	// GetTeamById returns a as_team as_event profile from event_as_teams.
	GetTeamById(ctx context.Context, id uuid.UUID) (*model.EventAsTeam, error)

	// GetPollById returns a poll as_event profile from event_as_polls.
	GetPollById(ctx context.Context, id uuid.UUID) (*model.EventAsPoll, error)

	// GetTestById returns a as_test as_event profile from event_as_tests.
	GetTestById(ctx context.Context, id uuid.UUID) (*model.EventAsTest, error)

	// GetLocationById returns coordinates for an as_event if its profile stores location data.
	GetLocationById(ctx context.Context, id uuid.UUID) (*events.EventView, error)

	// GetParticipants returns participants of the specified as_event.
	GetParticipants(ctx context.Context, id uuid.UUID) ([]model.EventParticipant, error)

	// Create stores a new classic as_event profile.
	Create(ctx context.Context, event *events.EventView) (*events.EventView, error)

	// Update updates as_event fields and optionally replaces organizers and available roles.
	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*events.EventView, error)

	// Delete removes an as_event profile by identifier.
	Delete(ctx context.Context, id uuid.UUID) error
}
