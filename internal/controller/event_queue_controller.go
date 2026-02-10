package controller

import (
	"ne_noy/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type eventQueueController struct {
	service service.EventQueueService
}

// getAllPosts godoc
//
//	@Summary	Получение всех постов в очереди событий
//	@Tags		EventQueue
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{array}		dto.EventQueueDto
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/queue [get]
//	@Security	VkAuth
func (ec *eventQueueController) getAllPosts(c *gin.Context) {
	posts, err := ec.service.GetAll(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, posts)
}

// createEventFromPost godoc
//
//	@Summary	Создание события из поста в очереди
//	@Tags		EventQueue
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/queue [post]
//	@Security	VkAuth
func (ec *eventQueueController) createEventFromPost(c *gin.Context) {
	c.Status(http.StatusOK)
}

// deletePostFromQueue godoc
//
//	@Summary	Удаление поста из очереди событий
//	@Tags		EventQueue
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header		string	true	"X-Request-Id"
//	@Success	200				{string}	string	"OK"
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	500				{object}	dto.ErrorResponse
//	@Router		/v1/events/queue [delete]
//	@Security	VkAuth
func (ec *eventQueueController) deletePostFromQueue(c *gin.Context) {
	c.Status(http.StatusOK)
}

func ConfigureEventQueueController(router *gin.RouterGroup, service service.EventQueueService) {
	ec := &eventQueueController{service: service}
	router.GET("/events/queue", ec.getAllPosts)
	router.POST("/events/queue", ec.createEventFromPost)
	router.DELETE("/events/queue", ec.deletePostFromQueue)
}
