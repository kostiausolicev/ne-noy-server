package event_as_test

import (
	"context"
	"errors"
	"ne_noy/internal/dto/test_dto"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_test"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

type eventTestService struct {
	repo repository.EventTestRepository
}

type EventTestService interface {
	// GetTest получение теста
	GetTest(ctx context.Context, testID uuid.UUID) (test_dto.TestDto, error)
	// GetQuestion получение конкретного вопроса
	GetQuestion(ctx context.Context, testID, questionID uuid.UUID) (test_dto.QuestionDto, error)
	// SetAnswer Установить ответ пользователем на вопрос
	SetAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.SetAnswerDto) (test_dto.UserAnswerDto, error)
	// AddQuestion Добавление вопроса в тест
	AddQuestion(ctx context.Context, testID uuid.UUID, question test_dto.AddQuestionDto) (test_dto.QuestionDto, error)
	// AddAnswer Добавление варианта ответа в тест
	AddAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.AddAnswerDto) (test_dto.AnswerDto, error)
	// CreateTest Создание теста
	CreateTest(ctx context.Context, test test_dto.CreateTestDto) (test_dto.TestDto, error)
	// UpdateTest Обновление теста
	UpdateTest(ctx context.Context, testID uuid.UUID, test test_dto.UpdateTestDto) (test_dto.TestDto, error)
}

func NewEventTestService(repo repository.EventTestRepository) EventTestService {
	return &eventTestService{repo: repo}
}

func (e *eventTestService) GetTest(ctx context.Context, testID uuid.UUID) (test_dto.TestDto, error) {
	test, err := e.repo.GetTest(ctx, testID)
	if err != nil {
		return test_dto.TestDto{}, err
	}

	return testToDto(*test), nil
}

func (e *eventTestService) GetQuestion(ctx context.Context, testID, questionID uuid.UUID) (test_dto.QuestionDto, error) {
	question, err := e.repo.GetQuestion(ctx, testID, questionID)
	if err != nil {
		return test_dto.QuestionDto{}, err
	}

	return questionToDto(*question), nil
}

func (e *eventTestService) SetAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.SetAnswerDto) (test_dto.UserAnswerDto, error) {
	if answer.UserID == uuid.Nil {
		return test_dto.UserAnswerDto{}, errors.New("user id is required")
	}
	if answer.AnswerID == nil && answer.Text == nil {
		return test_dto.UserAnswerDto{}, errors.New("answer id or text is required")
	}

	userAnswer, err := e.repo.SetUserAnswer(ctx, as_test.UserAnswer{
		UserID:     answer.UserID,
		QuestionID: questionID,
		AnswerID:   answer.AnswerID,
		Text:       answer.Text,
	})
	if err != nil {
		return test_dto.UserAnswerDto{}, err
	}

	return userAnswerToDto(*userAnswer), nil
}

func (e *eventTestService) AddQuestion(ctx context.Context, testID uuid.UUID, question test_dto.AddQuestionDto) (test_dto.QuestionDto, error) {
	if question.Text == "" {
		return test_dto.QuestionDto{}, errors.New("question text is required")
	}
	if question.Type == "" {
		return test_dto.QuestionDto{}, errors.New("question type is required")
	}

	createdQuestion, err := e.repo.AddQuestion(ctx, testID, as_test.Question{
		Text:   question.Text,
		Type:   question.Type,
		QOrder: question.Order,
	})
	if err != nil {
		return test_dto.QuestionDto{}, err
	}

	return questionToDto(*createdQuestion), nil
}

func (e *eventTestService) AddAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.AddAnswerDto) (test_dto.AnswerDto, error) {
	if answer.Text == "" {
		return test_dto.AnswerDto{}, errors.New("answer text is required")
	}

	createdAnswer, err := e.repo.AddAnswer(ctx, questionID, as_test.Answer{
		Text:      answer.Text,
		IsCorrect: answer.IsCorrect,
		Points:    answer.Points,
	})
	if err != nil {
		return test_dto.AnswerDto{}, err
	}

	return answerToDto(*createdAnswer), nil
}

func (e *eventTestService) CreateTest(ctx context.Context, test test_dto.CreateTestDto) (test_dto.TestDto, error) {
	if test.Name == "" {
		return test_dto.TestDto{}, errors.New("test name is required")
	}
	if test.Status == "" {
		return test_dto.TestDto{}, errors.New("test status is required")
	}
	if test.StartsAt.IsZero() {
		return test_dto.TestDto{}, errors.New("test starts_at is required")
	}

	attempts := test.Attempts
	if attempts == 0 {
		// В базе по умолчанию одна попытка, но явно задаем значение в модели для единообразного DTO.
		attempts = 1
	}

	createdTest, err := e.repo.CreateTest(ctx, &as_test.AsTest{
		EventProfile: events.EventProfile{
			Name:        test.Name,
			Description: test.Description,
			Cover:       test.Cover,
			Status:      test.Status,
			StartsAt:    test.StartsAt,
			EndsAt:      test.EndsAt,
		},
		ExtLinkID: test.ExtLinkID,
		Attempts:  attempts,
		VkPostID:  test.VkPostID,
	})
	if err != nil {
		return test_dto.TestDto{}, err
	}

	return testToDto(*createdTest), nil
}

func (e *eventTestService) UpdateTest(ctx context.Context, testID uuid.UUID, test test_dto.UpdateTestDto) (test_dto.TestDto, error) {
	update := as_test.AsTest{}
	if test.Name != nil {
		update.Name = *test.Name
	}
	update.Description = test.Description
	update.Cover = test.Cover
	if test.Status != nil {
		update.Status = *test.Status
	}
	if test.StartsAt != nil {
		update.StartsAt = *test.StartsAt
	}
	update.EndsAt = test.EndsAt
	update.ExtLinkID = test.ExtLinkID
	if test.Attempts != nil {
		update.Attempts = *test.Attempts
	}
	update.VkPostID = test.VkPostID

	updatedTest, err := e.repo.UpdateTest(ctx, testID, update)
	if err != nil {
		return test_dto.TestDto{}, err
	}

	return testToDto(*updatedTest), nil
}

func testToDto(test as_test.AsTest) test_dto.TestDto {
	questions := make([]test_dto.QuestionDto, 0, len(test.Questions))
	for _, question := range test.Questions {
		if question == nil {
			continue
		}
		questions = append(questions, questionToDto(*question))
	}

	return test_dto.TestDto{
		ID:          test.ID,
		Name:        test.Name,
		Description: test.Description,
		Cover:       test.Cover,
		Status:      test.Status,
		StartsAt:    test.StartsAt,
		EndsAt:      test.EndsAt,
		ExtLinkID:   test.ExtLinkID,
		Attempts:    test.Attempts,
		VkPostID:    test.VkPostID,
		Questions:   questions,
	}
}

func questionToDto(question as_test.Question) test_dto.QuestionDto {
	answers := make([]test_dto.AnswerDto, 0, len(question.Answers))
	for _, answer := range question.Answers {
		if answer == nil {
			continue
		}
		answers = append(answers, answerToDto(*answer))
	}

	return test_dto.QuestionDto{
		ID:      question.ID,
		Text:    question.Text,
		Type:    question.Type,
		EventID: question.EventID,
		Order:   question.QOrder,
		Answers: answers,
	}
}

func answerToDto(answer as_test.Answer) test_dto.AnswerDto {
	return test_dto.AnswerDto{
		ID:         answer.ID,
		QuestionID: answer.QuestionID,
		Text:       answer.Text,
		IsCorrect:  answer.IsCorrect,
		Points:     answer.Points,
	}
}

func userAnswerToDto(answer as_test.UserAnswer) test_dto.UserAnswerDto {
	return test_dto.UserAnswerDto{
		ID:         answer.ID,
		UserID:     answer.UserID,
		QuestionID: answer.QuestionID,
		AnswerID:   answer.AnswerID,
		Text:       answer.Text,
		Points:     answer.Points,
	}
}
