package service

import (
	"context"
	"encoding/json"
	"ne_noy/internal/dto"
	callback_dto "ne_noy/internal/dto/callback_dto"
	"ne_noy/internal/repository"
)

// EventQueueService предоставляет операции для работы с очередью постов
type EventQueueService interface {
	GetAll(ctx context.Context) ([]dto.EventQueueDto, error)
	DeletePostFromQueue(ctx context.Context, postID int64) error
}

type eventQueueService struct {
	repo repository.EventQueueRepository
}

func NewEventQueueService(repo repository.EventQueueRepository) EventQueueService {
	return &eventQueueService{repo: repo}
}

func (s *eventQueueService) GetAll(ctx context.Context) ([]dto.EventQueueDto, error) {
	models, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.EventQueueDto, 0, len(models))
	for _, m := range models {
		var pollDto dto.PollDto
		var attachmentsDto []dto.AttachmentDto
		var photosDto []dto.PhotoDto
		// Poll
		if len(m.Poll) > 0 {
			var polls []callback_dto.PollObject
			if err := json.Unmarshal(m.Poll, &polls); err == nil && len(polls) > 0 {
				p := polls[0]
				answers := make([]dto.AnswerDto, len(p.Answers))
				for i, a := range p.Answers {
					answers[i] = dto.AnswerDto{ID: a.ID, Text: a.Text}
				}
				pollDto = dto.PollDto{ID: p.ID, Answers: answers}
			}
		}

		// Attachments
		if len(m.Attachments) > 0 {
			var docs []callback_dto.DocObject
			if err := json.Unmarshal(m.Attachments, &docs); err == nil {
				for _, d := range docs {
					attachmentsDto = append(attachmentsDto, dto.AttachmentDto{ID: d.ID, Url: d.Url, Title: d.Title})
				}
			}
		}

		if len(m.Photos) > 0 {
			var photos []callback_dto.PhotoObject
			if err := json.Unmarshal(m.Photos, &photos); err == nil {
				for _, p := range photos {
					photosDto = append(photosDto, dto.PhotoDto{ID: p.ID, Sizes: func() []dto.PhotoSizeDto {
						sizes := make([]dto.PhotoSizeDto, len(p.Sizes))
						for i, s := range p.Sizes {
							sizes[i] = dto.PhotoSizeDto{Type: s.Type, Url: s.Url}
						}
						return sizes
					}()})
				}
			}
		}

		item := dto.EventQueueDto{
			PostID:      m.PostID,
			Text:        m.Text,
			Lat:         m.Lat,
			Lon:         m.Lon,
			Address:     m.Address,
			Poll:        pollDto,
			Photos:      photosDto,
			Attachments: attachmentsDto,
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *eventQueueService) DeletePostFromQueue(ctx context.Context, postID int64) error {
	return s.repo.RemovePostFromQueue(ctx, postID)
}
