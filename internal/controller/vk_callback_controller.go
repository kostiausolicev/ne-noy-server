package controller

import (
	"errors"
	"ne_noy/internal/dto/callback_dto"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type vkCallBackController struct {
	secret  string
	service service.VkCallbackService
}

const vkConfirmationResponse = "0bf0212a"

func (cb vkCallBackController) confirmHandler(c *gin.Context) {
	_, err := c.Writer.Write([]byte(vkConfirmationResponse))
	if err != nil {
		c.Error(err)
		return
	}
}

func (cb vkCallBackController) postNewHandler(c *gin.Context, dto callback_dto.GroupEvent) {
	var newPost callback_dto.NewPostEvent
	err := mapstructure.Decode(dto.Object.(map[string]interface{}), &newPost)
	if err != nil {
		c.Error(err)
		return
	}
	err = cb.service.AddPostToQueue(c.Request.Context(), newPost)
	if err != nil {
		c.Error(err)
		return
	}
	c.Writer.Write([]byte(vkConfirmationResponse))
	c.Status(200)
}

func (cb vkCallBackController) pollVoteNew(c *gin.Context, dto callback_dto.GroupEvent) {
	var newVote callback_dto.PollVoteNewDto
	err := mapstructure.Decode(dto.Object.(map[string]interface{}), &newVote)
	if err != nil {
		c.Error(err)
		return
	}
	err = cb.service.ApplyVote(c.Request.Context(), newVote)
	if err != nil {
		c.Error(err)
		return
	}
	c.Writer.Write([]byte(vkConfirmationResponse))
	c.Status(200)
}

func (cb vkCallBackController) handleConfirm(c *gin.Context) {
	var callback callback_dto.GroupEvent
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
	case "confirmation":
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
