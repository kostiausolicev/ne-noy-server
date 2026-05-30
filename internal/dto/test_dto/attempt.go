package test_dto

import (
	"time"

	"github.com/google/uuid"
)

type UserAttemptDto struct {
	ID            uuid.UUID `json:"id"`
	AttemptNumber int       `json:"attempt_number"`
	Points        int       `json:"points"`
	OrderNumber   int       `json:"order_number"`
}

type UserAttemptCreatedDto struct {
	ID            uuid.UUID `json:"id"`
	AttemptNumber int       `json:"attempt_number"`
	Started       time.Time `json:"started"`
}

type UserAttemptDetailDto struct {
	Info           UserAttemptDto `json:"info"`
	Answers        []AnswerDto    `json:"answers"`
	SelectedAnswer []uuid.UUID    `json:"selected_answer"`
}
