package controller

import (
	"io"
	"net/http"

	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
)

type healthImportController struct {
	service service.HealthImportService
}

// parseAppleMetadataZip godoc
//
//	@Summary		Распарсить метаданные тренировок из zip-файла
//	@Description	Принимает zip-архив c экспортом данных здоровья и возвращает распарсенные метаданные тренировок.
//	@Tags			health-import
//	@Accept			mpfd
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			platform		formData	string	true	"Платформа источника данных"
//	@Param			archive			formData	file	true	"Zip-архив с метаданными"
//	@Success		200				{object}	dto.UserActivitiesInfo
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/health-import/apple/parse [post]
//	@Security		VkAuth
func (hc *healthImportController) parseAppleMetadataZip(c *gin.Context) {
	vkID, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	file, err := c.FormFile("archive")
	if err != nil {
		c.Error(err)
		return
	}

	src, err := file.Open()
	if err != nil {
		c.Error(err)
		return
	}
	defer src.Close()

	archiveBytes, err := io.ReadAll(src)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := hc.service.ParceAppleMetadataZip(c.Request.Context(), vkID, dto.AppleArchiveZipDto{
		Platform: c.PostForm("platform"),
		Archive:  archiveBytes,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// saveUserActivities godoc
//
//	@Summary		Сохранить данные тренировки пользователя для события
//	@Description	Сохраняет тренировки текущего пользователя в рамках конкретного события.
//	@Tags			health-import events
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string					true	"Уникальный идентификатор запроса"
//	@Param			id				path		string					true	"UUID события"
//	@Param			payload			body		dto.UserActivitiesInfo	true	"Данные тренировки"
//	@Success		201
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/{id}/health-import [post]
//	@Security		VkAuth
func (hc *healthImportController) saveUserActivities(c *gin.Context) {
	eventID, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	vkID, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	var payload dto.UserActivitiesInfo
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(err)
		return
	}

	if err := hc.service.SaveAppleMetadata(c.Request.Context(), vkID, eventID, &payload); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusCreated)
}

// getUserActivities godoc
//
//	@Summary		Получить данные пользователя по тренировкам в событии
//	@Description	Возвращает сохранённые данные текущего пользователя по тренировкам в рамках конкретного события.
//	@Tags			health-import events
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			id				path		string	true	"UUID события"
//	@Success		200				{array}		dto.UserActivitiesInfo
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/events/{id}/health-import [get]
//	@Security		VkAuth
func (hc *healthImportController) getUserActivities(c *gin.Context) {
	eventID, err := ParseUUID(c, "id")
	if err != nil {
		c.Error(err)
		return
	}

	vkID, err := GetCtxInt64(c, config.UserVkIdContextKey)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := hc.service.GetUserActivities(c.Request.Context(), vkID, eventID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func ConfigureHealthImportController(router *gin.RouterGroup, service service.HealthImportService) {
	hc := &healthImportController{service: service}
	router.POST("/health-import/apple/parse", hc.parseAppleMetadataZip)
	router.POST("/events/:id/health-import", hc.saveUserActivities)
	router.GET("/events/:id/health-import", hc.getUserActivities)
}
