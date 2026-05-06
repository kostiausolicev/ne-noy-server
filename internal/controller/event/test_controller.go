package event

import (
	"fmt"
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

func (c *testController) CreateTest(ctx *gin.Context) {}

func (c *testController) GetTest(ctx *gin.Context) {}

func (c *testController) GetQuestion(ctx *gin.Context) {}

func (c *testController) SetAnswer(ctx *gin.Context) {}

func ConfigureTestController(
	router *gin.Engine,
	eventService event.EventService,
	testService event_as_test.EventTestService,
) {
	controller := &testController{
		eventService: eventService,
		testService:  testService,
	}

	router.GET(fmt.Sprintf("/events/test/:%s", idConst), controller.GetTest)
	router.GET(fmt.Sprintf("/events/test/:%s/q/:%s", idConst, questionIdConst), controller.GetQuestion)
	router.POST(fmt.Sprintf("/events/test/:%s/q/:%s", idConst, questionIdConst), controller.SetAnswer)
}
