package controller

import (
	"ne_noy/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type eventQueueController struct {
	service service.EventQueueService
}

const (
	routeEventQueue    = "/events/queue"
	routeQueueByPostID = "/queue/:postId"
)

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
//	@Param		X-Request-Id	header	string	true	"X-Request-Id"
//	@Success	200
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	500	{object}	dto.ErrorResponse
//	@Router		/v1/events/queue [post]
//	@Security	VkAuth
func (ec *eventQueueController) createEventFromPost(c *gin.Context) {
	c.Status(http.StatusOK)
}

// deletePostFromQueue godoc
//
//	@Summary	Удаление поста из очереди событий по ID поста
//	@Tags		EventQueue
//	@Accept		json
//	@Produce	json
//	@Param		X-Request-Id	header	string	true	"X-Request-Id"
//	@Param		postId			path	integer	true	"ID поста VK"
//	@Success	200
//	@Failure	400	{object}	dto.ErrorResponse	"Некорректный ID поста"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	500	{object}	dto.ErrorResponse
//	@Router		/v1/queue/{postId} [delete]
//	@Security	VkAuth
func (ec *eventQueueController) deletePostFromQueue(c *gin.Context) {
	postID, err := ParseInt64Param(c, ParamPostID)
	if err != nil {
		c.Error(err)
		return
	}

	if err = ec.service.DeletePostFromQueue(c.Request.Context(), postID); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

func ConfigureEventQueueController(router *gin.RouterGroup, service service.EventQueueService) {
	ec := &eventQueueController{service: service}
	router.GET(routeEventQueue, ec.getAllPosts)
	router.POST(routeEventQueue, ec.createEventFromPost)
	router.DELETE(routeQueueByPostID, ec.deletePostFromQueue)
}
