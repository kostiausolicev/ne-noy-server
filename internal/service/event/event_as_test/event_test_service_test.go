package event_as_test

import (
	"context"
	"errors"
	"ne_noy/internal/dto/test_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_test"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventTestServiceCreateTestSetsDefaultAttempts(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEventTestRepo()
	service := NewEventTestService(repo)

	test, err := service.CreateTest(ctx, test_dto.CreateTestDto{
		Name:     "Go Basics",
		Status:   "active",
		StartsAt: time.Now().UTC(),
	})

	require.NoError(t, err)
	require.Equal(t, "Go Basics", test.Name)
	require.Equal(t, 1, test.Attempts)
}

func TestEventTestServiceAddQuestionAndAnswerMapsDto(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	repo := newFakeEventTestRepo()
	repo.tests[testID] = &as_test.AsTest{EventProfile: events.EventProfile{BaseModel: model.BaseModel{ID: testID}}}
	service := NewEventTestService(repo)

	question, err := service.AddQuestion(ctx, testID, test_dto.AddQuestionDto{
		Text:  "Choose one",
		Type:  "single_choice",
		Order: 2,
	})
	require.NoError(t, err)
	require.Equal(t, testID, question.EventID)
	require.Equal(t, 2, question.Order)

	answer, err := service.AddAnswer(ctx, question.ID, test_dto.AddAnswerDto{
		Text:      "Correct",
		IsCorrect: true,
		Points:    10,
	})
	require.NoError(t, err)
	require.Equal(t, question.ID, answer.QuestionID)
	require.True(t, answer.IsCorrect)
	require.Equal(t, 10, answer.Points)
}

func TestEventTestServiceSetAnswerValidatesInputAndReturnsPoints(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.New()
	repo := newFakeEventTestRepo()
	repo.answers[answerID] = &as_test.Answer{
		BaseModel:  model.BaseModel{ID: answerID},
		QuestionID: questionID,
		Points:     7,
	}
	service := NewEventTestService(repo)

	_, err := service.SetAnswer(ctx, questionID, test_dto.SetAnswerDto{UserID: userID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "answer id or text is required")

	userAnswer, err := service.SetAnswer(ctx, questionID, test_dto.SetAnswerDto{
		UserID:   userID,
		AnswerID: &answerID,
	})
	require.NoError(t, err)
	require.Equal(t, userID, userAnswer.UserID)
	require.Equal(t, 7, userAnswer.Points)
}

func TestEventTestServiceGetTestMapsNestedQuestions(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.New()
	repo := newFakeEventTestRepo()
	repo.tests[testID] = &as_test.AsTest{
		EventProfile: events.EventProfile{
			BaseModel: model.BaseModel{ID: testID},
			Name:      "Nested",
			Status:    "active",
			StartsAt:  time.Now().UTC(),
		},
		Attempts: 1,
		Questions: []*as_test.Question{
			{
				BaseModel: model.BaseModel{ID: questionID},
				EventID:   testID,
				Text:      "Question",
				Type:      "single_choice",
				QOrder:    1,
				Answers: []*as_test.Answer{
					{BaseModel: model.BaseModel{ID: answerID}, QuestionID: questionID, Text: "Answer"},
				},
			},
		},
	}
	service := NewEventTestService(repo)

	test, err := service.GetTest(ctx, testID)
	require.NoError(t, err)
	require.Len(t, test.Questions, 1)
	require.Len(t, test.Questions[0].Answers, 1)
	require.Equal(t, "Answer", test.Questions[0].Answers[0].Text)
}

func TestEventTestServiceDeleteTestValidatesAndCallsRepository(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	repo := newFakeEventTestRepo()
	repo.tests[testID] = &as_test.AsTest{
		EventProfile: events.EventProfile{BaseModel: model.BaseModel{ID: testID}},
	}
	service := NewEventTestService(repo)

	err := service.DeleteTest(ctx, test_dto.DeleteTestDto{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "test id is required")

	require.NoError(t, service.DeleteTest(ctx, test_dto.DeleteTestDto{ID: testID}))
	_, err = repo.GetTest(ctx, testID)
	require.Error(t, err)
}

type fakeEventTestRepo struct {
	tests       map[uuid.UUID]*as_test.AsTest
	questions   map[uuid.UUID]*as_test.Question
	answers     map[uuid.UUID]*as_test.Answer
	userAnswers map[uuid.UUID]*as_test.UserAnswer
}

func newFakeEventTestRepo() *fakeEventTestRepo {
	return &fakeEventTestRepo{
		tests:       make(map[uuid.UUID]*as_test.AsTest),
		questions:   make(map[uuid.UUID]*as_test.Question),
		answers:     make(map[uuid.UUID]*as_test.Answer),
		userAnswers: make(map[uuid.UUID]*as_test.UserAnswer),
	}
}

func (f *fakeEventTestRepo) GetTest(_ context.Context, testID uuid.UUID) (*as_test.AsTest, error) {
	test, ok := f.tests[testID]
	if !ok {
		return nil, errors.New("test not found")
	}
	return test, nil
}

func (f *fakeEventTestRepo) GetQuestion(_ context.Context, testID, questionID uuid.UUID) (*as_test.Question, error) {
	question, ok := f.questions[questionID]
	if !ok || question.EventID != testID {
		return nil, errors.New("question not found")
	}
	return question, nil
}

func (f *fakeEventTestRepo) CreateTest(_ context.Context, test *as_test.AsTest) (*as_test.AsTest, error) {
	test.ID = uuid.New()
	test.CreatedAt = time.Now().UTC()
	f.tests[test.ID] = test
	return test, nil
}

func (f *fakeEventTestRepo) UpdateTest(_ context.Context, testID uuid.UUID, update as_test.AsTest) (*as_test.AsTest, error) {
	test, ok := f.tests[testID]
	if !ok {
		return nil, errors.New("test not found")
	}
	if update.Name != "" {
		test.Name = update.Name
	}
	if update.Status != "" {
		test.Status = update.Status
	}
	if update.Attempts != 0 {
		test.Attempts = update.Attempts
	}
	return test, nil
}

func (f *fakeEventTestRepo) DeleteTest(_ context.Context, testID uuid.UUID) error {
	if _, ok := f.tests[testID]; !ok {
		return errors.New("test not found")
	}
	delete(f.tests, testID)
	for questionID, question := range f.questions {
		if question.EventID == testID {
			delete(f.questions, questionID)
		}
	}
	return nil
}

func (f *fakeEventTestRepo) AddQuestion(_ context.Context, testID uuid.UUID, question as_test.Question) (*as_test.Question, error) {
	if _, ok := f.tests[testID]; !ok {
		return nil, errors.New("test not found")
	}
	question.ID = uuid.New()
	question.EventID = testID
	f.questions[question.ID] = &question
	return &question, nil
}

func (f *fakeEventTestRepo) AddAnswer(_ context.Context, questionID uuid.UUID, answer as_test.Answer) (*as_test.Answer, error) {
	if _, ok := f.questions[questionID]; !ok {
		return nil, errors.New("question not found")
	}
	answer.ID = uuid.New()
	answer.QuestionID = questionID
	f.answers[answer.ID] = &answer
	return &answer, nil
}

func (f *fakeEventTestRepo) SetUserAnswer(_ context.Context, userAnswer as_test.UserAnswer) (*as_test.UserAnswer, error) {
	userAnswer.ID = uuid.New()
	if userAnswer.AnswerID != nil {
		answer, ok := f.answers[*userAnswer.AnswerID]
		if !ok {
			return nil, errors.New("answer not found")
		}
		userAnswer.Points = answer.Points
	}
	f.userAnswers[userAnswer.ID] = &userAnswer
	return &userAnswer, nil
}
