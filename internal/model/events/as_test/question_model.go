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
	QuestionID   uuid.UUID
	Question     Question
	AttachmentID *int64
	Attachment   *model.Attachment
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
	UserID uuid.UUID
	User   model.User

	QuestionID uuid.UUID
	Question   Question

	AnswerID *uuid.UUID
	Answer   *Answer

	Text   *string
	Points int
}

func (q Question) TableName() string {
	return "questions"
}

func (q QuestionAttachment) TableName() string {
	return "question_attachments"
}

func (a Answer) TableName() string {
	return "answers"
}

func (u UserAnswer) TableName() string {
	return "user_answers"
}
