package event_as_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/test_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_test"
	"ne_noy/internal/repository"
	"strconv"

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
	// SetAnswer устанавливает ответ пользователя на вопрос
	SetAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.SetAnswerDto) (test_dto.UserAnswerDto, error)
	// UpdateAnswer обновляет существующий ответ пользователя на вопрос
	UpdateAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.UpdateAnswerDto) (test_dto.UserAnswerDto, error)
	// AddQuestion Добавление вопроса в тест
	AddQuestion(ctx context.Context, testID uuid.UUID, question test_dto.AddQuestionDto) (test_dto.QuestionDto, error)
	// UpdateQuestion обновляет текст, тип и порядок вопроса
	UpdateQuestion(ctx context.Context, testID, questionID uuid.UUID, question test_dto.AddQuestionDto) (test_dto.QuestionDto, error)
	// AddAnswer Добавление варианта ответа в тест
	AddAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.AddAnswerDto) (test_dto.AnswerDto, error)
	// CreateTest Создание теста
	CreateTest(ctx context.Context, test test_dto.CreateTestDto) (test_dto.TestDto, error)
	// UpdateTest Обновление теста
	UpdateTest(ctx context.Context, testID uuid.UUID, test test_dto.UpdateTestDto) (test_dto.TestDto, error)
	// DeleteTest удаляет тест
	DeleteTest(ctx context.Context, test test_dto.DeleteTestDto) error
	// GetMyTestResults возвращает ответы текущего пользователя на вопросы теста.
	// attemptID опционален: если передан — только ответы из этой попытки.
	GetMyTestResults(ctx context.Context, eventID, userID uuid.UUID, attemptID *uuid.UUID) ([]test_dto.MyTestResultDto, error)
	// GetUserTestResults возвращает результаты всех пользователей по тесту
	GetUserTestResults(ctx context.Context, eventID uuid.UUID) ([]test_dto.UserTestResultDto, error)
	// GetTestUsersDetail возвращает детальную информацию о пользователях теста: попытки, доступные ответы и выборки
	GetTestUsersDetail(ctx context.Context, testID uuid.UUID) ([]test_dto.TestUserResultDetailDto, error)
	// GenerateTestReport генерирует CSV-отчёт по результатам теста
	GenerateTestReport(ctx context.Context, eventID uuid.UUID) (test_dto.TestReportDto, error)
	// CreateAttempt создаёт новую попытку пользователя для теста
	CreateAttempt(ctx context.Context, userID, testID uuid.UUID) (test_dto.UserAttemptCreatedDto, error)
	// GetUserAttempts возвращает список попыток пользователя с баллами и порядковыми номерами
	GetUserAttempts(ctx context.Context, userID, testID uuid.UUID) ([]test_dto.UserAttemptDto, error)
}

func NewEventTestService(repo repository.EventTestRepository) EventTestService {
	return &eventTestService{repo: repo}
}

func attachmentDtosToModels(dtos []dto.AttachmentDto) []events.EventAttachment {
	result := make([]events.EventAttachment, 0, len(dtos))
	for _, d := range dtos {
		id := d.ID
		result = append(result, events.EventAttachment{
			AttachmentID: &id,
			Attachment: &model.Attachment{
				ID:       d.ID,
				Url:      d.Url,
				Filename: d.Title,
			},
		})
	}
	return result
}

func attachmentsToDto(attachments []events.EventAttachment) []dto.AttachmentDto {
	result := make([]dto.AttachmentDto, 0, len(attachments))
	for _, a := range attachments {
		if a.Attachment == nil || a.AttachmentID == nil {
			continue
		}
		result = append(result, dto.AttachmentDto{
			ID:    a.Attachment.ID,
			Url:   a.Attachment.Url,
			Title: a.Attachment.Filename,
		})
	}
	return result
}

func (e *eventTestService) GetTest(ctx context.Context, testID uuid.UUID) (test_dto.TestDto, error) {
	test, err := e.repo.GetTest(ctx, testID)
	if err != nil {
		return test_dto.TestDto{}, err
	}

	organizers, err := e.repo.GetEventOrganizers(ctx, testID)
	if err != nil {
		return test_dto.TestDto{}, err
	}
	test.Orgs = organizers

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
		AttemptID:  answer.AttemptID,
		Text:       answer.Text,
	})
	if err != nil {
		return test_dto.UserAnswerDto{}, err
	}

	return userAnswerToDto(*userAnswer), nil
}

func (e *eventTestService) UpdateAnswer(ctx context.Context, questionID uuid.UUID, answer test_dto.UpdateAnswerDto) (test_dto.UserAnswerDto, error) {
	if answer.UserID == uuid.Nil {
		return test_dto.UserAnswerDto{}, errors.New("user id is required")
	}
	if answer.AnswerID == nil && answer.Text == nil {
		return test_dto.UserAnswerDto{}, errors.New("answer id or text is required")
	}

	userAnswer, err := e.repo.UpdateUserAnswer(ctx, as_test.UserAnswer{
		UserID:     answer.UserID,
		QuestionID: questionID,
		AnswerID:   answer.AnswerID,
		AttemptID:  answer.AttemptID,
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

func (e *eventTestService) UpdateQuestion(ctx context.Context, testID, questionID uuid.UUID, question test_dto.AddQuestionDto) (test_dto.QuestionDto, error) {
	updated, err := e.repo.UpdateQuestion(ctx, testID, questionID, as_test.Question{
		Text:   question.Text,
		Type:   question.Type,
		QOrder: question.Order,
	})
	if err != nil {
		return test_dto.QuestionDto{}, err
	}

	return questionToDto(*updated), nil
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
			StartsAt:    test.StartsAt.Time,
			EndsAt:      test.EndsAt.ToTimePtr(),
		},
		EventRelations: events.EventRelations{
			Attachments:        attachmentDtosToModels(test.Attachments),
			AvailableRoleCodes: test.AvailableRoles,
		},
		ExtLinkID: test.ExtLinkID,
		Attempts:  attempts,
		VkPostID:  test.VkPostID,
	})
	if err != nil {
		return test_dto.TestDto{}, err
	}

	if len(test.Organizers) > 0 {
		if err = e.repo.SetEventOrganizers(ctx, createdTest.ID, test.Organizers); err != nil {
			return test_dto.TestDto{}, err
		}
	}

	organizers, err := e.repo.GetEventOrganizers(ctx, createdTest.ID)
	if err != nil {
		return test_dto.TestDto{}, err
	}
	createdTest.Orgs = organizers

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
		update.StartsAt = test.StartsAt.Time
	}
	update.EndsAt = test.EndsAt.ToTimePtr()
	update.ExtLinkID = test.ExtLinkID
	if test.Attempts != nil {
		update.Attempts = *test.Attempts
	}
	update.VkPostID = test.VkPostID
	if test.Attachments != nil {
		update.Attachments = attachmentDtosToModels(*test.Attachments)
	}
	update.AvailableRoleCodes = test.AvailableRoles

	updatedTest, err := e.repo.UpdateTest(ctx, testID, update)
	if err != nil {
		return test_dto.TestDto{}, err
	}

	if test.Organizers != nil {
		if err = e.repo.SetEventOrganizers(ctx, testID, *test.Organizers); err != nil {
			return test_dto.TestDto{}, err
		}
	}

	organizers, err := e.repo.GetEventOrganizers(ctx, testID)
	if err != nil {
		return test_dto.TestDto{}, err
	}
	updatedTest.Orgs = organizers

	return testToDto(*updatedTest), nil
}

func (e *eventTestService) DeleteTest(ctx context.Context, test test_dto.DeleteTestDto) error {
	if test.ID == uuid.Nil {
		return errors.New("test id is required")
	}

	// Сервис принимает DTO удаления, чтобы контроллер не зависел от модели репозитория.
	return e.repo.DeleteTest(ctx, test.ID)
}

func testToDto(test as_test.AsTest) test_dto.TestDto {
	questions := make([]test_dto.QuestionDto, 0, len(test.Questions))
	for _, question := range test.Questions {
		if question == nil {
			continue
		}
		questions = append(questions, questionToDto(*question))
	}

	orgs := make([]dto.UserMiniDto, 0, len(test.Orgs))
	for _, org := range test.Orgs {
		orgs = append(orgs, testUserToMiniDto(org))
	}

	roles := test.AvailableRoleCodes
	if roles == nil {
		roles = []string{}
	}

	return test_dto.TestDto{
		ID:             test.ID,
		Name:           test.Name,
		Description:    test.Description,
		Cover:          test.Cover,
		Status:         test.Status,
		StartsAt:       test.StartsAt,
		EndsAt:         test.EndsAt,
		ExtLinkID:      test.ExtLinkID,
		Attempts:       test.Attempts,
		VkPostID:       test.VkPostID,
		AvailableRoles: roles,
		Organizers:     orgs,
		Attachments:    attachmentsToDto(test.Attachments),
		Questions:      questions,
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

func (e *eventTestService) GetMyTestResults(ctx context.Context, eventID, userID uuid.UUID, attemptID *uuid.UUID) ([]test_dto.MyTestResultDto, error) {
	test, err := e.repo.GetTest(ctx, eventID)
	if err != nil {
		return nil, err
	}

	userAnswers, err := e.repo.GetUserAnswersByEvent(ctx, eventID, userID, attemptID)
	if err != nil {
		return nil, err
	}

	answersByQuestion := make(map[uuid.UUID][]string)
	for _, ua := range userAnswers {
		if ua.AnswerID != nil {
			answersByQuestion[ua.QuestionID] = append(answersByQuestion[ua.QuestionID], ua.AnswerID.String())
		}
	}

	result := make([]test_dto.MyTestResultDto, 0, len(test.Questions))
	for _, q := range test.Questions {
		if q == nil {
			continue
		}
		ids := answersByQuestion[q.ID]
		if ids == nil {
			ids = []string{}
		}
		result = append(result, test_dto.MyTestResultDto{
			Question:          questionToDto(*q),
			SelectedAnswerIds: ids,
		})
	}

	return result, nil
}

func (e *eventTestService) GetUserTestResults(ctx context.Context, eventID uuid.UUID) ([]test_dto.UserTestResultDto, error) {
	allAnswers, err := e.repo.GetAllUserAnswersByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	type userStats struct {
		user         model.User
		correctCount int
		totalCount   int
	}
	statsMap := make(map[uuid.UUID]*userStats)

	for _, ua := range allAnswers {
		stats, exists := statsMap[ua.UserID]
		if !exists {
			stats = &userStats{user: ua.User}
			statsMap[ua.UserID] = stats
		}
		stats.totalCount++
		if ua.Answer != nil && ua.Answer.IsCorrect {
			stats.correctCount++
		}
	}

	result := make([]test_dto.UserTestResultDto, 0, len(statsMap))
	for _, stats := range statsMap {
		result = append(result, test_dto.UserTestResultDto{
			User: testUserToMiniDto(stats.user),
			Attempts: []test_dto.TestAttemptDto{
				{
					CorrectCount: stats.correctCount,
					TotalCount:   stats.totalCount,
				},
			},
		})
	}

	return result, nil
}

func (e *eventTestService) GetTestUsersDetail(ctx context.Context, testID uuid.UUID) ([]test_dto.TestUserResultDetailDto, error) {
	test, err := e.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	allAnswers := make([]test_dto.AnswerDto, 0)
	for _, q := range test.Questions {
		if q == nil {
			continue
		}
		for _, a := range q.Answers {
			if a != nil {
				allAnswers = append(allAnswers, answerToDto(*a))
			}
		}
	}

	userAttempts, err := e.repo.GetTestUserAttempts(ctx, testID)
	if err != nil {
		return nil, err
	}

	type userEntry struct {
		user     model.User
		attempts []test_dto.UserAttemptDetailDto
	}
	order := make([]uuid.UUID, 0)
	entries := make(map[uuid.UUID]*userEntry)

	for _, ua := range userAttempts {
		if _, exists := entries[ua.UserID]; !exists {
			order = append(order, ua.UserID)
			entries[ua.UserID] = &userEntry{user: ua.User}
		}
		entries[ua.UserID].attempts = append(entries[ua.UserID].attempts, test_dto.UserAttemptDetailDto{
			Info: test_dto.UserAttemptDto{
				ID:            ua.ID,
				AttemptNumber: ua.AttemptNumber,
				Points:        ua.Points,
				OrderNumber:   ua.OrderNumber,
			},
			Answers:        allAnswers,
			SelectedAnswer: ua.SelectedAnswers,
		})
	}

	result := make([]test_dto.TestUserResultDetailDto, 0, len(order))
	for _, userID := range order {
		e := entries[userID]
		result = append(result, test_dto.TestUserResultDetailDto{
			User:     testUserToMiniDto(e.user),
			Attempts: e.attempts,
		})
	}

	return result, nil
}

func (e *eventTestService) GenerateTestReport(ctx context.Context, eventID uuid.UUID) (test_dto.TestReportDto, error) {
	test, err := e.repo.GetTest(ctx, eventID)
	if err != nil {
		return test_dto.TestReportDto{}, err
	}

	allAnswers, err := e.repo.GetAllUserAnswersByEvent(ctx, eventID)
	if err != nil {
		return test_dto.TestReportDto{}, err
	}

	questionMap := make(map[uuid.UUID]*as_test.Question)
	answerMap := make(map[uuid.UUID]*as_test.Answer)
	for _, q := range test.Questions {
		if q == nil {
			continue
		}
		questionMap[q.ID] = q
		for _, a := range q.Answers {
			if a != nil {
				answerMap[a.ID] = a
			}
		}
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"user_id", "question", "selected_answer", "is_correct", "points"})

	for _, ua := range allAnswers {
		questionText := ""
		if q, ok := questionMap[ua.QuestionID]; ok {
			questionText = q.Text
		}
		answerText := ""
		isCorrect := "false"
		if ua.Text != nil {
			answerText = *ua.Text
		} else if ua.AnswerID != nil {
			if a, ok := answerMap[*ua.AnswerID]; ok {
				answerText = a.Text
				if a.IsCorrect {
					isCorrect = "true"
				}
			}
		}
		_ = writer.Write([]string{
			ua.UserID.String(),
			questionText,
			answerText,
			isCorrect,
			strconv.Itoa(ua.Points),
		})
	}
	writer.Flush()

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return test_dto.TestReportDto{
		DownloadURL: "data:text/csv;base64," + encoded,
	}, nil
}

func (e *eventTestService) CreateAttempt(ctx context.Context, userID, testID uuid.UUID) (test_dto.UserAttemptCreatedDto, error) {
	attempts, err := e.repo.GetUserAttempts(ctx, userID, testID)
	if err != nil {
		return test_dto.UserAttemptCreatedDto{}, err
	}

	attempt, err := e.repo.CreateAttempt(ctx, userID, testID)
	if err != nil {
		return test_dto.UserAttemptCreatedDto{}, err
	}

	return test_dto.UserAttemptCreatedDto{
		ID:            attempt.ID,
		AttemptNumber: len(attempts) + 1,
		Started:       attempt.Started,
	}, nil
}

func (e *eventTestService) GetUserAttempts(ctx context.Context, userID, testID uuid.UUID) ([]test_dto.UserAttemptDto, error) {
	attempts, err := e.repo.GetUserAttempts(ctx, userID, testID)
	if err != nil {
		return nil, err
	}

	result := make([]test_dto.UserAttemptDto, 0, len(attempts))
	for _, a := range attempts {
		result = append(result, test_dto.UserAttemptDto{
			ID:            a.ID,
			AttemptNumber: a.AttemptNumber,
			Points:        a.Points,
			OrderNumber:   a.OrderNumber,
		})
	}
	return result, nil
}

func testUserToMiniDto(user model.User) dto.UserMiniDto {
	return dto.UserMiniDto{
		ID:        user.ID,
		VkId:      user.VkID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		PhotoURL:  user.PhotoURL,
	}
}
