package as_test

import (
	"ne_noy/internal/model"
	"time"

	"github.com/google/uuid"
)

type UserAttempt struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	TestID  uuid.UUID
	Started time.Time
}

type UserAttemptInfo struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TestID        uuid.UUID
	Started       time.Time
	Points        int
	AttemptNumber int
	OrderNumber   int
}

type UserAttemptWithSelections struct {
	UserAttemptInfo
	User            model.User
	SelectedAnswers []uuid.UUID
}

func (u UserAttempt) TableName() string {
	return "user_attempts"
}

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

	AttemptID *uuid.UUID

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
