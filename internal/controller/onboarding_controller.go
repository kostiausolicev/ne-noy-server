package controller

import (
	"ne_noy/internal/config"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
)

type onboardingController struct {
	service service.OnboardingService
}

// getOnboardings godoc
//
//	@Summary		Получить список онбордингов для пользователя
//	@Description	Получить список онбордингов для пользователя. Онбординги - это этапы, которые пользователь должен пройти для полного использования функционала приложения. Например, онбординг может включать в себя заполнение профиля, добавление друзей, участие в мероприятиях и т.д.
//	@Tags			onboarding
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса для трассировки"
//	@Param			platform		query		string	true	"Платформа, для которой запрашиваются онбординги."
//	@Success		200				{array}		dto.OnboardingDto
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/onboardings/all [get]
//	@Security		VkAuth
func (oc *onboardingController) getOnboardings(c *gin.Context) {
	platform := c.Query("platform")

	onboardings, err := oc.service.GetAllOnboardingCodesByPlatform(c.Request.Context(), platform)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, onboardings)
}

// getOnboardingsForUser godoc
//
//	@Summary		Получить список онбордингов для пользователя
//	@Description	Получить список онбордингов для пользователя. Онбординги - это этапы, которые пользователь должен пройти для полного использования функционала приложения. Например, онбординг может включать в себя заполнение профиля, добавление друзей, участие в мероприятиях и т.д.
//	@Tags			onboarding
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса для трассировки"
//	@Param			platform		query		string	true	"Платформа, для которой запрашиваются онбординги."
//	@Success		200				{array}		dto.OnboardingDto
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/onboardings [get]
//	@Security		VkAuth
func (oc *onboardingController) getOnboardingsForUser(c *gin.Context) {
	vkId, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}
	platform := c.Query("platform")

	onboardings, err := oc.service.GetOnboardingsForUser(c.Request.Context(), vkId, platform)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, onboardings)
}

// setUserOnboarding godoc
//
//	@Summary		Пометить онбординг как пройденный для пользователя
//	@Description	Пометить онбординг как пройденный для пользователя. Это означает, что пользователь прошел определенный этап онбординга и может перейти к следующему этапу или использовать функционал, который был заблокирован до прохождения этого этапа.
//	@Tags			onboarding
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса для трассировки"
//	@Param			X-Request-Id	path	string	true	"Уникальный идентификатор онбординга."
//	@Success		201
//	@Failure		401	{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/onboardings/:id [post]
//	@Security		VkAuth
func (oc *onboardingController) setUserOnboarding(c *gin.Context) {
	vkId, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}
	onboardingID := c.Param("id")
	err = oc.service.SetUserOnboarding(c.Request.Context(), vkId, onboardingID)
	if err != nil {
		c.Error(err)
		return
	}
}

func ConfigureOnboardingController(router *gin.RouterGroup, service service.OnboardingService) {
	oc := &onboardingController{service: service}
	router.GET("/onboardings", oc.getOnboardingsForUser)
	router.GET("/onboardings/all", oc.getOnboardings)
	router.POST("/onboardings/:id", oc.setUserOnboarding)
}
