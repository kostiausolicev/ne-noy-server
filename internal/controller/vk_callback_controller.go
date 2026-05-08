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

const (
	routeVKCallback = "/callback"

	vkConfirmationResponse = "0bf0212a"
	vkEventConfirmation    = "confirmation"
	vkEventWallPostNew     = "wall_post_new"
	vkEventPollVoteNew     = "poll_vote_new"
)

var errInvalidCallbackSecret = errors.New("secret does not match")

func (cb vkCallBackController) confirmHandler(c *gin.Context) {
	writeCallbackOK(c)
}

func (cb vkCallBackController) postNewHandler(c *gin.Context, dto callback_dto.GroupEvent) {
	newPost, ok := decodeCallbackObject[callback_dto.NewPostEvent](c, dto)
	if !ok {
		return
	}
	err := cb.service.AddPostToQueue(c.Request.Context(), newPost)
	if err != nil {
		c.Error(err)
		return
	}
	writeCallbackOK(c)
}

func (cb vkCallBackController) pollVoteNew(c *gin.Context, dto callback_dto.GroupEvent) {
	newVote, ok := decodeCallbackObject[callback_dto.PollVoteNewDto](c, dto)
	if !ok {
		return
	}
	err := cb.service.ApplyVote(c.Request.Context(), newVote)
	if err != nil {
		c.Error(err)
		return
	}
	writeCallbackOK(c)
}

func (cb vkCallBackController) handleConfirm(c *gin.Context) {
	callback, ok := BindJSON[callback_dto.GroupEvent](c)
	if !ok {
		return
	}
	if callback.Secret != cb.secret {
		c.Error(errInvalidCallbackSecret)
		return
	}
	switch callback.EventType {
	case vkEventConfirmation:
		cb.confirmHandler(c)
	case vkEventWallPostNew:
		cb.postNewHandler(c, callback)
	case vkEventPollVoteNew:
		cb.pollVoteNew(c, callback)
	}
}

func decodeCallbackObject[T any](c *gin.Context, dto callback_dto.GroupEvent) (T, bool) {
	var payload T
	object, ok := dto.Object.(map[string]interface{})
	if !ok {
		c.Error(ParseError)
		return payload, false
	}
	if err := mapstructure.Decode(object, &payload); err != nil {
		c.Error(err)
		return payload, false
	}
	return payload, true
}

func writeCallbackOK(c *gin.Context) {
	if _, err := c.Writer.Write([]byte(vkConfirmationResponse)); err != nil {
		c.Error(err)
		return
	}
	c.Status(200)
}

func ConfigureVkCallBackController(router *gin.RouterGroup, secret string, service service.VkCallbackService) {
	controller := &vkCallBackController{secret: secret, service: service}

	router.POST(routeVKCallback, controller.handleConfirm)
}
