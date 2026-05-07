package repository

import (
	"context"
	"ne_noy/internal/model/events/as_test"

	"github.com/google/uuid"
)

type EventTestRepository interface {
	// GetTest возвращает тест вместе с вопросами и вариантами ответов.
	GetTest(ctx context.Context, testID uuid.UUID) (*as_test.AsTest, error)

	// GetQuestion возвращает один вопрос теста вместе с вариантами ответов.
	GetQuestion(ctx context.Context, testID, questionID uuid.UUID) (*as_test.Question, error)

	// CreateTest создает профиль мероприятия-теста.
	CreateTest(ctx context.Context, test *as_test.AsTest) (*as_test.AsTest, error)

	// UpdateTest обновляет поля профиля мероприятия-теста.
	UpdateTest(ctx context.Context, testID uuid.UUID, update as_test.AsTest) (*as_test.AsTest, error)

	// AddQuestion добавляет вопрос в тест.
	AddQuestion(ctx context.Context, testID uuid.UUID, question as_test.Question) (*as_test.Question, error)

	// AddAnswer добавляет вариант ответа к вопросу.
	AddAnswer(ctx context.Context, questionID uuid.UUID, answer as_test.Answer) (*as_test.Answer, error)

	// SetUserAnswer сохраняет ответ пользователя и начисленные баллы.
	SetUserAnswer(ctx context.Context, userAnswer as_test.UserAnswer) (*as_test.UserAnswer, error)
}
