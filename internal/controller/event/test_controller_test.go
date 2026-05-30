package event_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ne_noy/internal/config"
	"ne_noy/internal/controller/event"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/test_dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func authMiddleware(vkID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(config.UserVkIdContextKey, vkID)
		c.Next()
	}
}

func TestGetMyTestResultsReturns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware(12345))
	testSvc := &fakeTestService{myResults: []test_dto.MyTestResultDto{
		{Question: test_dto.QuestionDto{Text: "Q1"}, SelectedAnswerIds: []string{}},
	}}
	event.ConfigureTestController(
		router.Group("/"),
		&fakeEventService{},
		testSvc,
		&fakeUserService{userID: uuid.New()},
	)

	eventID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events/"+eventID.String()+"/test/my-results", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var result []test_dto.MyTestResultDto
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Len(t, result, 1)
	require.Equal(t, "Q1", result[0].Question.Text)
}

func TestGetUserTestResultsReturns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware(12345))
	userID := uuid.New()
	testSvc := &fakeTestService{userResults: []test_dto.UserTestResultDto{
		{
			User:     dto.UserMiniDto{ID: userID, FirstName: "Ivan"},
			Attempts: []test_dto.TestAttemptDto{{CorrectCount: 3, TotalCount: 5}},
		},
	}}
	event.ConfigureTestController(
		router.Group("/"),
		&fakeEventService{},
		testSvc,
		&fakeUserService{userID: userID},
	)

	eventID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events/"+eventID.String()+"/test/user-results", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var result []test_dto.UserTestResultDto
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Len(t, result, 1)
	require.Equal(t, 3, result[0].Attempts[0].CorrectCount)
}

func TestGenerateTestReportReturns200WithDataURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware(12345))
	testSvc := &fakeTestService{report: test_dto.TestReportDto{DownloadURL: "data:text/csv;base64,abc"}}
	event.ConfigureTestController(
		router.Group("/"),
		&fakeEventService{},
		testSvc,
		&fakeUserService{userID: uuid.New()},
	)

	eventID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events/"+eventID.String()+"/test/report", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var result test_dto.TestReportDto
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Contains(t, result.DownloadURL, "data:text/csv")
}

func TestGetMyTestResultsReturns400OnInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware(12345))
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": c.Errors.Last().Error()})
		}
	})
	event.ConfigureTestController(
		router.Group("/"),
		&fakeEventService{},
		&fakeTestService{},
		&fakeUserService{userID: uuid.New()},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events/not-a-uuid/test/my-results", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// --- fakes ---

type fakeTestService struct {
	myResults   []test_dto.MyTestResultDto
	userResults []test_dto.UserTestResultDto
	report      test_dto.TestReportDto
}

func (f *fakeTestService) GetTest(_ context.Context, _ uuid.UUID) (test_dto.TestDto, error) {
	return test_dto.TestDto{}, nil
}
func (f *fakeTestService) GetQuestion(_ context.Context, _, _ uuid.UUID) (test_dto.QuestionDto, error) {
	return test_dto.QuestionDto{}, nil
}
func (f *fakeTestService) SetAnswer(_ context.Context, _ uuid.UUID, _ test_dto.SetAnswerDto) (test_dto.UserAnswerDto, error) {
	return test_dto.UserAnswerDto{}, nil
}
func (f *fakeTestService) UpdateAnswer(_ context.Context, _ uuid.UUID, _ test_dto.UpdateAnswerDto) (test_dto.UserAnswerDto, error) {
	return test_dto.UserAnswerDto{}, nil
}
func (f *fakeTestService) AddQuestion(_ context.Context, _ uuid.UUID, _ test_dto.AddQuestionDto) (test_dto.QuestionDto, error) {
	return test_dto.QuestionDto{}, nil
}
func (f *fakeTestService) UpdateQuestion(_ context.Context, _, _ uuid.UUID, _ test_dto.AddQuestionDto) (test_dto.QuestionDto, error) {
	return test_dto.QuestionDto{}, nil
}
func (f *fakeTestService) AddAnswer(_ context.Context, _ uuid.UUID, _ test_dto.AddAnswerDto) (test_dto.AnswerDto, error) {
	return test_dto.AnswerDto{}, nil
}
func (f *fakeTestService) CreateTest(_ context.Context, _ test_dto.CreateTestDto) (test_dto.TestDto, error) {
	return test_dto.TestDto{}, nil
}
func (f *fakeTestService) UpdateTest(_ context.Context, _ uuid.UUID, _ test_dto.UpdateTestDto) (test_dto.TestDto, error) {
	return test_dto.TestDto{}, nil
}
func (f *fakeTestService) DeleteTest(_ context.Context, _ test_dto.DeleteTestDto) error { return nil }
func (f *fakeTestService) GetMyTestResults(_ context.Context, _, _ uuid.UUID, _ *uuid.UUID) ([]test_dto.MyTestResultDto, error) {
	return f.myResults, nil
}
func (f *fakeTestService) GetUserTestResults(_ context.Context, _ uuid.UUID) ([]test_dto.UserTestResultDto, error) {
	return f.userResults, nil
}
func (f *fakeTestService) GetTestUsersDetail(_ context.Context, _ uuid.UUID) ([]test_dto.TestUserResultDetailDto, error) {
	return nil, nil
}
func (f *fakeTestService) GenerateTestReport(_ context.Context, _ uuid.UUID) (test_dto.TestReportDto, error) {
	return f.report, nil
}
func (f *fakeTestService) CreateAttempt(_ context.Context, _, _ uuid.UUID) (test_dto.UserAttemptCreatedDto, error) {
	return test_dto.UserAttemptCreatedDto{}, nil
}
func (f *fakeTestService) GetUserAttempts(_ context.Context, _, _ uuid.UUID) ([]test_dto.UserAttemptDto, error) {
	return nil, nil
}

type fakeUserService struct {
	userID uuid.UUID
}

func (f *fakeUserService) GetUserByVkId(_ context.Context, _ int64) (*dto.UserDto, error) {
	id := f.userID
	return &dto.UserDto{ID: &id}, nil
}

func (f *fakeUserService) UpdateRole(_ context.Context, _ int64, _ uuid.UUID) error { return nil }
func (f *fakeUserService) GetAllUsers(_ context.Context, _ string) ([]dto.UserMiniDto, error) {
	return nil, nil
}
func (f *fakeUserService) UpdatePermissions(_ context.Context, _ string, _ int64, _ bool) error {
	return nil
}
func (f *fakeUserService) CreateUser(_ context.Context, _ dto.CreateUserDto) (*dto.UserDto, error) {
	return nil, nil
}
func (f *fakeUserService) CreateUserByLinks(_ context.Context, _ []string) ([]*dto.UserDto, error) {
	return nil, nil
}

type fakeEventService struct{}

func (f *fakeEventService) PublishEvent(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeEventService) GetAll(_ context.Context, _ int64) ([]dto.EventMiniDto, error) {
	return nil, nil
}
func (f *fakeEventService) GetEventsByRole(_ context.Context, _ string) ([]dto.EventMiniDto, error) {
	return nil, nil
}
func (f *fakeEventService) GetArchiveEvents(_ context.Context, _ string) ([]dto.EventMiniDto, error) {
	return nil, nil
}
