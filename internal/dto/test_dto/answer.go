package test_dto

import "github.com/google/uuid"

type AnswerDto struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `json:"text"`
	IsCorrect  bool      `json:"is_correct"`
	Points     int       `json:"points"`
}

type AddAnswerDto struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	Points    int    `json:"points"`
}

type SetAnswerDto struct {
	UserID    uuid.UUID  `json:"-"`
	AttemptID *uuid.UUID `json:"attempt_id"`
	AnswerID  *uuid.UUID `json:"answer_id"`
	Text      *string    `json:"text"`
}

type UpdateAnswerDto struct {
	UserID    uuid.UUID  `json:"-"`
	AttemptID *uuid.UUID `json:"attempt_id"`
	AnswerID  *uuid.UUID `json:"answer_id"`
	Text      *string    `json:"text"`
}

type UserAnswerDto struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	QuestionID uuid.UUID  `json:"question_id"`
	AnswerID   *uuid.UUID `json:"answer_id"`
	Text       *string    `json:"text"`
	Points     int        `json:"points"`
}
