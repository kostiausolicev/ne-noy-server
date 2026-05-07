package event

import (
	"fmt"
	"ne_noy/internal/controller"
	"ne_noy/internal/dto/test_dto"
	"ne_noy/internal/service/event"
	"ne_noy/internal/service/event/event_as_test"

	"github.com/gin-gonic/gin"
)

const (
	idConst         = "id"
	questionIdConst = "qId"
)

type testController struct {
	eventService event.EventService
	testService  event_as_test.EventTestService
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
	var createTestDto test_dto.CreateTestDto
	if err := ctx.ShouldBindJSON(&createTestDto); err != nil {
		ctx.Error(err)
		return
	}

	test, err := c.testService.CreateTest(ctx.Request.Context(), createTestDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, test)
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
	testID, err := controller.ParseUUID(ctx, idConst)
	if err != nil {
		ctx.Error(err)
		return
	}

	test, err := c.testService.GetTest(ctx.Request.Context(), testID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, test)
}

// GetQuestion godoc
//
//	@Summary	Получить вопрос теста
//	@Tags		tests
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
	testID, err := controller.ParseUUID(ctx, idConst)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, questionIdConst)
	if err != nil {
		ctx.Error(err)
		return
	}

	question, err := c.testService.GetQuestion(ctx.Request.Context(), testID, questionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, question)
}

// SetAnswer godoc
//
//	@Summary	Сохранить ответ пользователя на вопрос
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param		id				path		string					true	"UUID теста"
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
	testID, err := controller.ParseUUID(ctx, idConst)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, questionIdConst)
	if err != nil {
		ctx.Error(err)
		return
	}

	var setAnswerDto test_dto.SetAnswerDto
	if err = ctx.ShouldBindJSON(&setAnswerDto); err != nil {
		ctx.Error(err)
		return
	}

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

	ctx.JSON(200, answer)
}

// AddQuestion godoc
//
//	@Summary	Добавить вопрос в тест
//	@Tags		tests
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string						true	"Уникальный идентификатор запроса"
//	@Param		id				path		string						true	"UUID теста"
//	@Param		request			body		test_dto.AddQuestionDto	true	"Данные вопроса"
//	@Success	200				{object}	test_dto.QuestionDto
//	@Failure	400				{object}	dto.ErrorResponse	"Некорректные данные"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	404				{object}	dto.ErrorResponse	"Тест не найден"
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/test/{id}/q [post]
//	@Security	VkAuth
func (c *testController) AddQuestion(ctx *gin.Context) {
	testID, err := controller.ParseUUID(ctx, idConst)
	if err != nil {
		ctx.Error(err)
		return
	}

	var addQuestionDto test_dto.AddQuestionDto
	if err = ctx.ShouldBindJSON(&addQuestionDto); err != nil {
		ctx.Error(err)
		return
	}

	question, err := c.testService.AddQuestion(ctx.Request.Context(), testID, addQuestionDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, question)
}

// AddAnswer godoc
//
//	@Summary	Добавить вариант ответа к вопросу
//	@Tags		tests
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
	testID, err := controller.ParseUUID(ctx, idConst)
	if err != nil {
		ctx.Error(err)
		return
	}
	questionID, err := controller.ParseUUID(ctx, questionIdConst)
	if err != nil {
		ctx.Error(err)
		return
	}

	var addAnswerDto test_dto.AddAnswerDto
	if err = ctx.ShouldBindJSON(&addAnswerDto); err != nil {
		ctx.Error(err)
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

	ctx.JSON(200, answer)
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
	testID, err := controller.ParseUUID(ctx, idConst)
	if err != nil {
		ctx.Error(err)
		return
	}

	var updateTestDto test_dto.UpdateTestDto
	if err = ctx.ShouldBindJSON(&updateTestDto); err != nil {
		ctx.Error(err)
		return
	}

	test, err := c.testService.UpdateTest(ctx.Request.Context(), testID, updateTestDto)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, test)
}

func ConfigureTestController(
	router *gin.RouterGroup,
	eventService event.EventService,
	testService event_as_test.EventTestService,
) {
	c := &testController{
		eventService: eventService,
		testService:  testService,
	}

	router.POST("/events/test", c.CreateTest)
	router.GET(fmt.Sprintf("/events/test/:%s", idConst), c.GetTest)
	router.PATCH(fmt.Sprintf("/events/test/:%s", idConst), c.UpdateTest)
	router.POST(fmt.Sprintf("/events/test/:%s/q", idConst), c.AddQuestion)
	router.GET(fmt.Sprintf("/events/test/:%s/q/:%s", idConst, questionIdConst), c.GetQuestion)
	router.POST(fmt.Sprintf("/events/test/:%s/q/:%s", idConst, questionIdConst), c.SetAnswer)
	router.POST(fmt.Sprintf("/events/test/:%s/q/:%s/answers", idConst, questionIdConst), c.AddAnswer)
}
