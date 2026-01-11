package controller

import (
	"errors"
	"ne_noy/internal/dto/group_event"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type vkCallBackController struct {
	secret  string
	service service.VkCallbackService
}

func (cb vkCallBackController) confirmHandler(c *gin.Context) {
	_, err := c.Writer.Write([]byte("3460e36c"))
	if err != nil {
		c.Error(err)
		return
	}
}

func (cb vkCallBackController) postNewHandler(c *gin.Context, dto group_event.GroupEvent) {
	var newPost group_event.NewPostEvent
	err := mapstructure.Decode(dto.Object.(map[string]interface{}), &newPost)
	if err != nil {
		c.Error(err)
		return
	}
	err = cb.service.AddPostToQueue(newPost)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
	return
}

func (cb vkCallBackController) pollVoteNew(c *gin.Context, dto group_event.GroupEvent) {
	var newVote group_event.PollVoteNewDto
	err := mapstructure.Decode(dto.Object.(map[string]interface{}), &newVote)
	if err != nil {
		c.Error(err)
		return
	}
	err = cb.service.ApplyVote(newVote)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
}

func (cb vkCallBackController) handleConfirm(c *gin.Context) {
	var callback group_event.GroupEvent
	err := c.ShouldBindJSON(&callback)
	if err != nil {
		c.Error(err)
		return
	}
	if callback.Secret != cb.secret {
		c.Error(errors.New("secret does not match"))
		return
	}
	switch callback.EventType {
	case "confirm":
		cb.confirmHandler(c)
	case "wall_post_new":
		cb.postNewHandler(c, callback)
	case "poll_vote_new":
		cb.pollVoteNew(c, callback)
	}
}

func ConfigureVkCallBackController(router *gin.RouterGroup, secret string, service service.VkCallbackService) {
	controller := &vkCallBackController{secret: secret, service: service}

	router.POST("/callback", controller.handleConfirm)
}
