package as_test

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

type Question struct {
	model.BaseModel
	Text    string
	Type    string
	EventID uuid.UUID
	Event   AsTest
	QOrder  int

	Answers     []*Answer
	Attachments []*QuestionAttachment
}

type QuestionAttachment struct {
	model.BaseModel
	QuestionId uuid.UUID
	Question   Question
	Attachment model.Attachment
}

type Answer struct {
	model.BaseModel
	QuestionID uuid.UUID
	Question   Question
	IsCorrect  bool
	Text       string
	Points     int
}

type UserAnswer struct {
	model.BaseModel
	UserId uuid.UUID
	User   model.User

	QuestionId uuid.UUID
	Question   Question

	AnswerId uuid.UUID
	Answer   Answer

	Points int
}
