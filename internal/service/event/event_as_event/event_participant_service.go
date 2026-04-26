package event_as_event

import (
	"context"
	"errors"
	"math"
	"ne_noy/internal/apperror"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

type eventParticipantService struct {
	epr      repository.EventParticipantRepository
	er       repository.EventRepository
	distance int64
}

func NewEventParticipantService(epr repository.EventParticipantRepository, er repository.EventRepository, distance int64) EventParticipantService {
	return eventParticipantService{epr: epr, er: er, distance: distance}
}

type EventParticipantService interface {
	ParticipantToEvent(ctx context.Context, eventID uuid.UUID, userVkID int64, prepareType string) (bool, error)
	UnParticipantToEvent(ctx context.Context, eventID uuid.UUID, userVkID int64) (bool, error)
	CheckParticipant(ctx context.Context, participantData dto.CheckEventParticipant) error
}

func (eps eventParticipantService) ParticipantToEvent(ctx context.Context, eventID uuid.UUID, userVkID int64, prepareType string) (bool, error) {
	return eps.epr.Participant(ctx, eventID, userVkID, prepareType)
}

func (eps eventParticipantService) UnParticipantToEvent(ctx context.Context, eventID uuid.UUID, userVkID int64) (bool, error) {
	return eps.epr.UnParticipant(ctx, eventID, userVkID)
}

func (eps eventParticipantService) CheckParticipant(ctx context.Context, participantData dto.CheckEventParticipant) (err error) {
	switch participantData.CheckType {
	case "Personal QR":
		{
			err = eps.checkByAdmin(ctx, participantData)
		}
	case "Admin panel":
		{
			err = eps.checkByAdmin(ctx, participantData)
		}
	case "Event QR":
		{
			err = eps.checkByQr(ctx, participantData)
		}
	}
	return err
}

func (eps eventParticipantService) checkByAdmin(ctx context.Context, participantData dto.CheckEventParticipant) error {
	orgs, err := eps.er.GetEventOrgs(ctx, participantData.EventId)
	if err != nil {
		return err
	}
	for _, eventOrg := range orgs {
		if eventOrg.VkID == *participantData.CheckAuthorVkId {
			participant := model.EventParticipant{
				EventID:        participantData.EventId,
				UserID:         participantData.UserId,
				IsChecked:      true,
				CheckTimestamp: &participantData.Timestamp,
				CheckLat:       participantData.Lat,
				CheckLong:      participantData.Long,
				CheckType:      participantData.CheckType,
				CheckAuthor:    &eventOrg.ID,
			}
			err := eps.epr.CheckParticipant(ctx, &participant)
			return err
		}
	}
	return errors.New("participant not exist")
}

func (eps eventParticipantService) checkByQr(ctx context.Context, participantData dto.CheckEventParticipant) error {
	event, err := eps.er.GetLocationById(ctx, participantData.EventId)
	if err != nil {
		return err
	}
	if event.Lat == nil || event.Long == nil {
		return apperror.EventLocationNotSetErr
	}

	diff := haversineDistance(*event.Lat, *event.Long, *participantData.Lat, *participantData.Long)

	// Вынести в конфиги
	if diff > float64(eps.distance) {
		return apperror.ParticipantLocationTooLageErr
	}

	participant := model.EventParticipant{
		EventID:        participantData.EventId,
		UserID:         participantData.UserId,
		IsChecked:      true,
		CheckTimestamp: &participantData.Timestamp,
		CheckLat:       participantData.Lat,
		CheckLong:      participantData.Long,
		CheckType:      participantData.CheckType,
	}
	err = eps.epr.CheckParticipant(ctx, &participant)
	if errors.Is(err, apperror.ParticipantNotExistErr) {
		err = nil
		// Если участника не было, то
		_, err := eps.epr.ParticipantById(ctx, participantData.EventId, participantData.UserId, "app")
		if err != nil {
			return err
		}
		err = eps.epr.CheckParticipant(ctx, &participant)
		if err != nil {
			return err
		}
	}
	return err
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	// Средняя широта для коррекции долготы
	avgLat := (lat1Rad + lat2Rad) / 2

	// Разница координат в радианах
	dlat := (lat2 - lat1) * math.Pi / 180
	dlon := (lon2 - lon1) * math.Pi / 180

	// Корректировка долготы с учетом широты
	x := dlon * math.Cos(avgLat)
	y := dlat

	// Расстояние по теореме Пифагора
	return 6371.0 * math.Sqrt(x*x+y*y) * 1000
}
