package event

import (
	"net/http"

	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/dto/test_dto"
	appservice "ne_noy/internal/service"
	"ne_noy/internal/service/event"
	"ne_noy/internal/service/event/event_as_test"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type testController struct {
	eventService event.EventService
	testService  event_as_test.EventTestService
	userService  appservice.UserService
}

// CreateTest godoc
//
//	@Summary	Создать тест
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param		request			body		test_dto.CreateTestDto	true	"Данные для создания теста"
//	@Success	200				{object}	test_dto.TestDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test [post]
//	@Security	VkAuth
func (c *testController) CreateTest(ctx *gin.Context) {
	createTestDto, ok := controller.BindJSON[test_dto.CreateTestDto](ctx)
	if !ok {
		return
	}

	test, err := c.testService.CreateTest(ctx.Request.Context(), createTestDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, test)
}

// GetTest godoc
//
//	@Summary	Получить тест по ID
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		id				path		string	true	"UUID теста"
//	@Success	200				{object}	test_dto.TestDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id} [get]
//	@Security	VkAuth
func (c *testController) GetTest(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	test, err := c.testService.GetTest(ctx.Request.Context(), testID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, test)
}

// GetQuestion godoc
//
//	@Summary	Получить вопрос теста
//	@Tags		test-questions
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		id				path		string	true	"UUID теста"
//	@Param		qId				path		string	true	"UUID вопроса"
//	@Success	200				{object}	test_dto.QuestionDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Вопрос не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q/{qId} [get]
//	@Security	VkAuth
func (c *testController) GetQuestion(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, controller.ParamQuestionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	question, err := c.testService.GetQuestion(ctx.Request.Context(), testID, questionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, question)
}

// SetAnswer godoc
//
//	@Summary	Сохранить ответ пользователя на вопрос
//	@Tags		test-answers
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param		id				path		string						true	"UUID теста"
//	@Param		qId				path		string					true	"UUID вопроса"
//	@Param		request			body		test_dto.SetAnswerDto	true	"Ответ пользователя"
//	@Success	200				{object}	test_dto.UserAnswerDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Вопрос или вариант ответа не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q/{qId} [post]
//	@Security	VkAuth
func (c *testController) SetAnswer(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, controller.ParamQuestionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	setAnswerDto, ok := controller.BindJSON[test_dto.SetAnswerDto](ctx)
	if !ok {
		return
	}

	userID, err := c.currentUserID(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	setAnswerDto.UserID = userID

	// Проверяем принадлежность вопроса тесту из URL, чтобы нельзя было ответить на чужой вопрос через другой testID.
	if _, err = c.testService.GetQuestion(ctx.Request.Context(), testID, questionID); err != nil {
		ctx.Error(err)
		return
	}

	answer, err := c.testService.SetAnswer(ctx.Request.Context(), questionID, setAnswerDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, answer)
}

// UpdateAnswer godoc
//
//	@Summary	Обновить ответ пользователя на вопрос
//	@Tags		test-answers
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param		id				path		string						true	"UUID теста"
//	@Param		qId				path		string						true	"UUID вопроса"
//	@Param		request			body		test_dto.UpdateAnswerDto	true	"Новые данные ответа"
//	@Success	200				{object}	test_dto.UserAnswerDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Ответ не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q/{qId} [patch]
//	@Security	VkAuth
func (c *testController) UpdateAnswer(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, controller.ParamQuestionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	updateDto, ok := controller.BindJSON[test_dto.UpdateAnswerDto](ctx)
	if !ok {
		return
	}

	userID, err := c.currentUserID(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	updateDto.UserID = userID

	if _, err = c.testService.GetQuestion(ctx.Request.Context(), testID, questionID); err != nil {
		ctx.Error(err)
		return
	}

	answer, err := c.testService.UpdateAnswer(ctx.Request.Context(), questionID, updateDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, answer)
}

// AddQuestion godoc
//
//	@Summary	Добавить вопрос в тест
//	@Tags		test-questions
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param		id				path		string					true	"UUID теста"
//	@Param		request			body		test_dto.AddQuestionDto	true	"Данные вопроса"
//	@Success	200				{object}	test_dto.QuestionDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q [post]
//	@Security	VkAuth
func (c *testController) AddQuestion(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	addQuestionDto, ok := controller.BindJSON[test_dto.AddQuestionDto](ctx)
	if !ok {
		return
	}

	question, err := c.testService.AddQuestion(ctx.Request.Context(), testID, addQuestionDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, question)
}

// UpdateQuestion godoc
//
//	@Summary	Обновить данные вопроса
//	@Tags		test-questions
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param		id				path		string					true	"UUID теста"
//	@Param		qId				path		string					true	"UUID вопроса"
//	@Param		request			body		test_dto.AddQuestionDto	true	"Новые данные вопроса"
//	@Success	200				{object}	test_dto.QuestionDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Вопрос не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q/{qId}/info [patch]
//	@Security	VkAuth
func (c *testController) UpdateQuestion(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, controller.ParamQuestionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	updateDto, ok := controller.BindJSON[test_dto.AddQuestionDto](ctx)
	if !ok {
		return
	}

	question, err := c.testService.UpdateQuestion(ctx.Request.Context(), testID, questionID, updateDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, question)
}

// AddAnswer godoc
//
//	@Summary	Добавить вариант ответа к вопросу
//	@Tags		test-answers
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param		id				path		string					true	"UUID теста"
//	@Param		qId				path		string					true	"UUID вопроса"
//	@Param		request			body		test_dto.AddAnswerDto	true	"Данные варианта ответа"
//	@Success	200				{object}	test_dto.AnswerDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Вопрос не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q/{qId}/answers [post]
//	@Security	VkAuth
func (c *testController) AddAnswer(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, controller.ParamQuestionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	addAnswerDto, ok := controller.BindJSON[test_dto.AddAnswerDto](ctx)
	if !ok {
		return
	}

	// Проверяем связку тест-вопрос до добавления ответа, чтобы URL оставался источником контекста операции.
	if _, err = c.testService.GetQuestion(ctx.Request.Context(), testID, questionID); err != nil {
		ctx.Error(err)
		return
	}

	answer, err := c.testService.AddAnswer(ctx.Request.Context(), questionID, addAnswerDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, answer)
}

// UpdateTest godoc
//
//	@Summary	Обновить тест
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param		id				path		string					true	"UUID теста"
//	@Param		request			body		test_dto.UpdateTestDto	true	"Данные для обновления теста"
//	@Success	200				{object}	test_dto.TestDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id} [patch]
//	@Security	VkAuth
func (c *testController) UpdateTest(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	updateTestDto, ok := controller.BindJSON[test_dto.UpdateTestDto](ctx)
	if !ok {
		return
	}

	test, err := c.testService.UpdateTest(ctx.Request.Context(), testID, updateTestDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, test)
}

// DeleteTest godoc
//
//	@Summary	Удалить тест
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param		id				path	string	true	"UUID теста"
//	@Success	200
//	@Failure	400	{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500	{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id} [delete]
//	@Security	VkAuth
func (c *testController) DeleteTest(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err = c.testService.DeleteTest(ctx.Request.Context(), test_dto.DeleteTestDto{ID: testID}); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusOK)
}

// GetMyTestResults godoc
//
//	@Summary	Получить свои результаты теста
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		eventId			path		string	true	"UUID мероприятия-теста"
//	@Param		attemptId		query		string	false	"UUID попытки (фильтр по конкретной попытке)"
//	@Success	200				{array}		test_dto.MyTestResultDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{eventId}/test/my-results [get]
//	@Security	VkAuth
func (c *testController) GetMyTestResults(ctx *gin.Context) {
	eventID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	userID, err := c.currentUserID(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var attemptID *uuid.UUID
	if raw := ctx.Query("attemptId"); raw != "" {
		parsed, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			ctx.Error(controller.ParseError)
			return
		}
		attemptID = &parsed
	}

	results, err := c.testService.GetMyTestResults(ctx.Request.Context(), eventID, userID, attemptID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

// CreateAttempt godoc
//
//	@Summary	Начать новую попытку прохождения теста
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		eventId			path		string	true	"UUID мероприятия-теста"
//	@Success	200				{object}	test_dto.UserAttemptCreatedDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{eventId}/test/attempts [post]
//	@Security	VkAuth
func (c *testController) CreateAttempt(ctx *gin.Context) {
	eventID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	userID, err := c.currentUserID(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	attempt, err := c.testService.CreateAttempt(ctx.Request.Context(), userID, eventID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, attempt)
}

// GetUserAttempts godoc
//
//	@Summary	Получить список своих попыток прохождения теста
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		eventId			path		string	true	"UUID мероприятия-теста"
//	@Success	200				{array}		test_dto.UserAttemptDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{eventId}/test/attempts [get]
//	@Security	VkAuth
func (c *testController) GetUserAttempts(ctx *gin.Context) {
	eventID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	userID, err := c.currentUserID(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	attempts, err := c.testService.GetUserAttempts(ctx.Request.Context(), userID, eventID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, attempts)
}

// GetTestUsersDetail godoc
//
//	@Summary	Получить детальную информацию о пользователях теста
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		id				path		string	true	"UUID теста"
//	@Success	200				{array}		test_dto.TestUserResultDetailDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/users-detail [get]
//	@Security	VkAuth
func (c *testController) GetTestUsersDetail(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	result, err := c.testService.GetTestUsersDetail(ctx.Request.Context(), testID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetUserTestResults godoc
//
//	@Summary	Получить результаты всех пользователей по тесту
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		eventId			path		string	true	"UUID мероприятия-теста"
//	@Success	200				{array}		test_dto.UserTestResultDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{eventId}/test/user-results [get]
//	@Security	VkAuth
func (c *testController) GetUserTestResults(ctx *gin.Context) {
	eventID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	results, err := c.testService.GetUserTestResults(ctx.Request.Context(), eventID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

// GenerateTestReport godoc
//
//	@Summary	Сгенерировать CSV-отчёт по результатам теста
//	@Tags		old-tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param		eventId			path		string	true	"UUID мероприятия-теста"
//	@Success	200				{object}	test_dto.TestReportDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректный UUID"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/{eventId}/test/report [get]
//	@Security	VkAuth
func (c *testController) GenerateTestReport(ctx *gin.Context) {
	eventID, err := controller.ParseUUID(ctx, controller.ParamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	report, err := c.testService.GenerateTestReport(ctx.Request.Context(), eventID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, report)
}

func (c *testController) currentUserID(ctx *gin.Context) (uuid.UUID, error) {
	vkID, err := controller.GetCtxInt64(ctx, config.UserVkIdContextKey)
	if err != nil {
		return uuid.Nil, err
	}

	user, err := c.userService.GetUserByVkId(ctx.Request.Context(), vkID)
	if err != nil {
		return uuid.Nil, err
	}
	if user == nil || user.ID == nil {
		return uuid.Nil, controller.ParseError
	}

	return *user.ID, nil
}

func ConfigureTestController(
	router *gin.RouterGroup,
	eventService event.EventService,
	testService event_as_test.EventTestService,
	userService appservice.UserService,
) {
	c := &testController{
		eventService: eventService,
		testService:  testService,
		userService:  userService,
	}

	router.POST(routeTest, c.CreateTest)
	router.GET(routeTestByID, c.GetTest)
	router.PATCH(routeTestByID, c.UpdateTest)
	router.DELETE(routeTestByID, c.DeleteTest)
	router.POST(routeTestQuestion, c.AddQuestion)
	router.GET(routeTestQuestionByID, c.GetQuestion)
	router.POST(routeTestQuestionByID, c.SetAnswer)
	router.PATCH(routeTestQuestionByID, c.UpdateAnswer)
	router.POST(routeTestQuestionAnswers, c.AddAnswer)
	router.PATCH(routeTestQuestionInfo, c.UpdateQuestion)

	router.GET(routeTestUsersDetail, c.GetTestUsersDetail)
	router.GET(routeEventTestMyResults, c.GetMyTestResults)
	router.GET(routeEventTestUserResults, c.GetUserTestResults)
	router.GET(routeEventTestReport, c.GenerateTestReport)
	router.POST(routeEventTestAttempts, c.CreateAttempt)
	router.GET(routeEventTestAttempts, c.GetUserAttempts)
}
