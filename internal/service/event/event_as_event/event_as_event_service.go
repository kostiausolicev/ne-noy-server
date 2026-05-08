package event_as_event

import (
	"context"
	"errors"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/event_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/repository"
	"time"

	"github.com/google/uuid"
)

type eventAsEventService struct {
	repo repository.EventEventRepository
}

type EventAsEventService interface {
	GetEventById(ctx context.Context, eventId uuid.UUID) (event_dto.EventDto, error)
	UpdateEvent(ctx context.Context, eventId uuid.UUID, event event_dto.CreateUpdateEventDto) (event_dto.EventDto, error)
	CreateEvent(ctx context.Context, event event_dto.CreateUpdateEventDto) (event_dto.EventDto, error)
	GetEventParticipants(ctx context.Context, eventId uuid.UUID) ([]dto.UserMiniDto, error)
}

func NewEventAsEventService(repo repository.EventEventRepository) EventAsEventService {
	return &eventAsEventService{repo: repo}
}

func (e *eventAsEventService) GetEventById(ctx context.Context, eventId uuid.UUID) (event_dto.EventDto, error) {
	event, err := e.repo.GetEventById(ctx, eventId)
	if err != nil {
		return event_dto.EventDto{}, err
	}
	return asEventToDto(*event), nil
}

func (e *eventAsEventService) UpdateEvent(ctx context.Context, eventId uuid.UUID, event event_dto.CreateUpdateEventDto) (event_dto.EventDto, error) {
	fields := make(map[string]interface{})
	if event.VkPostId != nil {
		fields["vk_post_id"] = event.VkPostId
	}
	if event.VkVoteID != nil {
		fields["vk_vote_id"] = event.VkVoteID
	}
	if event.VkPollAnswerID != nil {
		fields["vk_poll_answer_id"] = event.VkPollAnswerID
	}
	if event.PhotoURL != nil {
		fields["cover"] = event.PhotoURL
	}
	if event.Title != nil {
		fields["name"] = event.Title
	}
	if event.Description != nil {
		fields["description"] = event.Description
	}
	if event.Address != nil {
		fields["address"] = event.Address
	}
	if event.AdAddress != nil {
		fields["additional_address"] = event.AdAddress
	}
	if event.Lat != nil {
		fields["lat"] = event.Lat
	}
	if event.Long != nil {
		fields["lon"] = event.Long
	}
	if event.Status != "" {
		fields["status"] = event.Status
	}
	if event.StartsAt != nil {
		fields["starts_at"] = event.StartsAt
	}
	if event.EndsAt != nil {
		fields["ends_at"] = event.EndsAt
	}

	var orgs []model.User
	if event.Orgs != nil {
		orgs = miniUsersToModels(event.Orgs)
	}
	updated, err := e.repo.Update(ctx, eventId, fields, orgs, nil)
	if err != nil {
		return event_dto.EventDto{}, err
	}
	return asEventToDto(*updated), nil
}

func (e *eventAsEventService) CreateEvent(ctx context.Context, event event_dto.CreateUpdateEventDto) (event_dto.EventDto, error) {
	if event.Title == nil || *event.Title == "" {
		return event_dto.EventDto{}, errors.New("event title is required")
	}
	if event.Status == "" {
		return event_dto.EventDto{}, errors.New("event status is required")
	}
	if event.StartsAt == nil || event.StartsAt.IsZero() {
		return event_dto.EventDto{}, errors.New("event startsAt is required")
	}
	if event.EndsAt == nil || event.EndsAt.IsZero() {
		return event_dto.EventDto{}, errors.New("event endsAt is required")
	}

	created, err := e.repo.CreateEvent(ctx, &as_event.AsEvent{
		EventProfile: modelEventProfile(
			*event.Title,
			event.Description,
			event.PhotoURL,
			event.Status,
			*event.StartsAt,
			event.EndsAt,
		),
		EventRelations:    modelEventRelations(event.Orgs),
		VkPostID:          event.VkPostId,
		VkVoteID:          event.VkVoteID,
		VkPollAnswerID:    event.VkPollAnswerID,
		Lat:               event.Lat,
		Lon:               event.Long,
		Address:           event.Address,
		AdditionalAddress: event.AdAddress,
	})
	if err != nil {
		return event_dto.EventDto{}, err
	}
	return asEventToDto(*created), nil
}

func (e *eventAsEventService) GetEventParticipants(ctx context.Context, eventId uuid.UUID) ([]dto.UserMiniDto, error) {
	participants, err := e.repo.GetParticipants(ctx, eventId, 0)
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserMiniDto, len(participants))
	for i, participant := range participants {
		result[i] = userToMiniDto(participant.User)
	}
	return result, nil
}

func modelEventProfile(name string, description, cover *string, status string, startsAt time.Time, endsAt *time.Time) events.EventProfile {
	return events.EventProfile{
		Name:        name,
		Description: description,
		Cover:       cover,
		Status:      status,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	}
}

func modelEventRelations(orgs []dto.UserMiniDto) events.EventRelations {
	return events.EventRelations{
		Orgs: miniUsersToModels(orgs),
	}
}

func miniUsersToModels(users []dto.UserMiniDto) []model.User {
	result := make([]model.User, 0, len(users))
	for _, user := range users {
		result = append(result, model.User{
			BaseModel: model.BaseModel{ID: user.ID},
			VkID:      user.VkId,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			PhotoURL:  user.PhotoURL,
		})
	}
	return result
}

func asEventToDto(event as_event.AsEvent) event_dto.EventDto {
	orgs := make([]dto.UserMiniDto, 0, len(event.Orgs))
	for _, org := range event.Orgs {
		orgs = append(orgs, userToMiniDto(org))
	}

	participants := make([]dto.UserMiniDto, 0, len(event.EventParticipants))
	for _, participant := range event.EventParticipants {
		participants = append(participants, userToMiniDto(participant.User))
	}

	return event_dto.EventDto{
		ID:                event.ID,
		VkPostId:          event.VkPostID,
		VkVoteID:          event.VkVoteID,
		VkPollAnswerID:    event.VkPollAnswerID,
		Lat:               event.Lat,
		Long:              event.Lon,
		PhotoURL:          event.Cover,
		Title:             event.Name,
		Description:       event.Description,
		Attachments:       make([]dto.AttachmentDto, 0),
		Orgs:              orgs,
		Address:           event.Address,
		AdAddress:         event.AdditionalAddress,
		Participants:      participants,
		ParticipantsCount: event.ParticipantsCount,
		Status:            event.Status,
		StartsAt:          event.StartsAt,
		EndsAt:            derefTimeOrZero(event.EndsAt),
	}
}

func userToMiniDto(user model.User) dto.UserMiniDto {
	return dto.UserMiniDto{
		ID:        user.ID,
		VkId:      user.VkID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		PhotoURL:  user.PhotoURL,
	}
}

func derefTimeOrZero(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
