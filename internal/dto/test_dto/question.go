package test_dto

import (
	"ne_noy/internal/dto"

	"github.com/google/uuid"
)

type QuestionDto struct {
	ID          uuid.UUID           `json:"id"`
	Text        string              `json:"text"`
	Type        string              `json:"type"`
	EventID     uuid.UUID           `json:"event_id"`
	Order       int                 `json:"order"`
	Attachments []dto.AttachmentDto `json:"attachments"`
	Answers     []AnswerDto         `json:"answers"`
}

type AddQuestionDto struct {
	Text  string `json:"text"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}
