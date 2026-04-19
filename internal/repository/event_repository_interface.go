package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// EventRepository describes persistence operations for events and event profiles.
type EventRepository interface {
	// GetAll returns all events visible in the system.
	GetAll(ctx context.Context) ([]*model.Event, error)

	// GetAllByOrg returns all events where the specified user is an organizer.
	GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*model.Event, error)

	// GetAllByRole returns active events available for the specified role.
	GetAllByRole(ctx context.Context, role string) ([]*model.Event, error)

	// GetAllArchive returns archived active events available for the specified role.
	GetAllArchive(ctx context.Context, role string) ([]*model.Event, error)

	// GetEventLocationData returns location data for an event by its identifier.
	GetEventLocationData(ctx context.Context, id uuid.UUID) (*model.Event, error)

	// GetEventTypeById returns the profile type of an event.
	GetEventTypeById(ctx context.Context, id uuid.UUID) (string, error)

	// GetUserParticipationInEvent checks whether a VK user participates in the event.
	GetUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userVkId int64) (bool, error)

	// GetEventOrgs returns organizers of the specified event.
	GetEventOrgs(ctx context.Context, eventId uuid.UUID) ([]model.User, error)

	// GetByVkPollId returns an event linked to the specified VK poll identifier.
	GetByVkPollId(ctx context.Context, pollId int64) (*model.Event, error)

	// GetEventById returns a classic event profile from event_as_events.
	GetEventById(ctx context.Context, id uuid.UUID) (*model.EventAsEvent, error)

	// GetActivityById returns an activity event profile from event_as_activities.
	GetActivityById(ctx context.Context, id uuid.UUID) (*model.EventAsActivity, error)

	// GetTeamById returns a team event profile from event_as_teams.
	GetTeamById(ctx context.Context, id uuid.UUID) (*model.EventAsTeam, error)

	// GetPollById returns a poll event profile from event_as_polls.
	GetPollById(ctx context.Context, id uuid.UUID) (*model.EventAsPoll, error)

	// GetTestById returns a test event profile from event_as_tests.
	GetTestById(ctx context.Context, id uuid.UUID) (*model.EventAsTest, error)

	// GetLocationById returns coordinates for an event if its profile stores location data.
	GetLocationById(ctx context.Context, id uuid.UUID) (*model.Event, error)

	// GetParticipants returns participants of the specified event.
	GetParticipants(ctx context.Context, id uuid.UUID) ([]model.EventParticipant, error)

	// Create stores a new classic event profile.
	Create(ctx context.Context, event *model.Event) (*model.Event, error)

	// Update updates event fields and optionally replaces organizers and available roles.
	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*model.Event, error)

	// Delete removes an event profile by identifier.
	Delete(ctx context.Context, id uuid.UUID) error
}
