package service

import (
	"context"
	"encoding/json"
	"ne_noy/internal/dto/callback_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type VkCallbackService interface {
	ApplyVote(ctx context.Context, dto callback_dto.PollVoteNewDto) error
	AddPostToQueue(ctx context.Context, dto callback_dto.NewPostEvent) error
}

type vkCallBackService struct {
	repository              repository.EventQueueRepository
	eventRepository         repository.EventRepository
	eventParticipantService EventParticipantService
}

// ApplyVote TODO надо сделать записи в отдельную таблицу, чтобы не терять данные до создания события
func (v vkCallBackService) ApplyVote(ctx context.Context, dto callback_dto.PollVoteNewDto) error {
	// TODO контекст падает, нужно разобраться
	event, err := v.eventRepository.GetByVkPollId(ctx, dto.PollID)
	if err != nil {
		return err
	}
	if event.VkPollAnswerID != nil && *event.VkPollAnswerID == dto.OptionID {
		_, err = v.eventParticipantService.ParticipantToEvent(ctx, event.ID, dto.UserID, "vote")
	} else if event.VkPollAnswerID != nil && *event.VkPollAnswerID != dto.OptionID {
		_, err = v.eventParticipantService.UnParticipantToEvent(ctx, event.ID, dto.UserID)
	}
	if err != nil {
		return err
	}
	return nil
}

func (v vkCallBackService) AddPostToQueue(ctx context.Context, dto callback_dto.NewPostEvent) error {
	// Оставляем только документы и получаем срез *DocObject (или DocObject)
	var attachments []callback_dto.DocObject
	var photos []callback_dto.PhotoObject
	var poll []callback_dto.PollObject

	for _, attachment := range dto.Attachments {
		if attachment.Type == "doc" {
			attachments = append(attachments, *attachment.Doc)
		}
		if attachment.Type == "photo" {
			photos = append(photos, *attachment.Photo)
		}
		if attachment.Type == "poll" {
			poll = append(poll, *attachment.Poll)
		}
	}

	jsonAttachments, err := json.Marshal(attachments)
	if err != nil {
		return err
	}
	jsonPhotos, err := json.Marshal(photos)
	if err != nil {
		return err
	}
	jsonPoll, err := json.Marshal(poll)
	if err != nil {
		return err
	}
	id, _ := uuid.NewUUID()
	eventQueueModel := model.EventQueueModel{
		ID:     id,
		PostID: dto.ID,
		Text:   dto.Text,
		//Lat:         dto.Geo.Place.Lat,
		//Lon:         dto.Geo.Place.Lon,
		//Address:     dto.Geo.Place.Address,
		Attachments: datatypes.JSON(jsonAttachments),
		Photos:      datatypes.JSON(jsonPhotos),
		Poll:        datatypes.JSON(jsonPoll),
	}

	err = v.repository.AddPostToQueue(ctx, &eventQueueModel)
	if err != nil {
		return err
	}

	return nil
}

func NewVkCallbackService(repository repository.EventQueueRepository, eventRepository repository.EventRepository, eventParticipantService EventParticipantService) VkCallbackService {
	return vkCallBackService{repository: repository, eventRepository: eventRepository, eventParticipantService: eventParticipantService}
}
