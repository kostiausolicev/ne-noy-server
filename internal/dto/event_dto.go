package dto

import (
	"time"

	"github.com/google/uuid"
)

type EventMiniDto struct {
	ID                uuid.UUID     `json:"id"`
	Title             string        `json:"title"`
	Orgs              []UserMiniDto `json:"orgs"`
	Participants      []UserMiniDto `json:"participants"`
	ParticipantsCount int           `json:"participantsCount"`
	Status            string        `json:"status"`
	StartsAt          time.Time     `json:"startAt"`
}
