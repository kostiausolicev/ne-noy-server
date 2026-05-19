package test_dto

import (
	"time"

	"github.com/google/uuid"
)

type TestDto struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Cover       *string       `json:"cover"`
	Status      string        `json:"status"`
	StartsAt    time.Time     `json:"starts_at"`
	EndsAt      *time.Time    `json:"ends_at"`
	ExtLinkID   *string       `json:"ext_link_id"`
	Attempts    int           `json:"attempts"`
	VkPostID    *int64        `json:"vk_post_id"`
	Questions   []QuestionDto `json:"questions"`
}

type CreateTestDto struct {
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Cover       *string    `json:"cover"`
	Status      string     `json:"status"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	ExtLinkID   *string    `json:"ext_link_id"`
	Attempts    int        `json:"attempts"`
	VkPostID    *int64     `json:"vk_post_id"`
}

type UpdateTestDto struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Cover       *string    `json:"cover"`
	Status      *string    `json:"status"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	ExtLinkID   *string    `json:"ext_link_id"`
	Attempts    *int       `json:"attempts"`
	VkPostID    *int64     `json:"vk_post_id"`
}

type DeleteTestDto struct {
	ID uuid.UUID `json:"id"`
}

type QuestionDto struct {
	ID      uuid.UUID   `json:"id"`
	Text    string      `json:"text"`
	Type    string      `json:"type"`
	EventID uuid.UUID   `json:"event_id"`
	Order   int         `json:"order"`
	Answers []AnswerDto `json:"answers"`
}

type AddQuestionDto struct {
	Text  string `json:"text"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}

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
	UserID   uuid.UUID  `json:"user_id"`
	AnswerID *uuid.UUID `json:"answer_id"`
	Text     *string    `json:"text"`
}

type UserAnswerDto struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	QuestionID uuid.UUID  `json:"question_id"`
	AnswerID   *uuid.UUID `json:"answer_id"`
	Text       *string    `json:"text"`
	Points     int        `json:"points"`
}
