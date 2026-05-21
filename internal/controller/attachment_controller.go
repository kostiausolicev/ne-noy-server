package controller

import (
	"net/http"
	"strconv"

	"ne_noy/internal/dto"
	"ne_noy/proto/gen"

	"github.com/gin-gonic/gin"
)

type attachmentController struct {
	client gen.AttachmentServiceClient
}

const (
	routeAttachments      = "/attachments"
	routeAttachmentsBatch = "/attachments/batch"
	routeAttachmentByID   = "/attachments/:id"
	routeAttachmentFile   = "/file"

	queryTTL   = "ttl"
	queryForce = "force"
	querySign  = "sign"

	defaultTTL = 3600
)

// putOne godoc
//
//	@Summary		Загрузить один файл
//	@Description	Передаёт файл в сервис вложений через gRPC и возвращает UUID созданной записи
//	@Tags			attachments
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string				true	"Уникальный идентификатор запроса"
//	@Param			body			body		dto.PutOneRequest	true	"Данные файла"
//	@Success		201				{object}	dto.PutOneResponse
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные входные данные"
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/attachments [post]
//	@Security		VkAuth
func (ac *attachmentController) putOne(c *gin.Context) {
	req, ok := BindJSON[dto.PutOneRequest](c)
	if !ok {
		return
	}
	resp, err := ac.client.PutOne(c.Request.Context(), &gen.PutOneRequest{
		File: &gen.UploadFileDTO{
			StorageType: gen.StorageType(req.File.StorageType),
			FileName:    req.File.FileName,
			FileContent: req.File.FileContent,
		},
		Async: req.Async,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.PutOneResponse{ID: resp.Id})
}

// putMany godoc
//
//	@Summary		Загрузить несколько файлов
//	@Description	Передаёт список файлов в сервис вложений через gRPC и возвращает список UUID
//	@Tags			attachments
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string				true	"Уникальный идентификатор запроса"
//	@Param			body			body		dto.PutManyRequest	true	"Список файлов"
//	@Success		201				{object}	dto.PutManyResponse
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/attachments/batch [post]
//	@Security		VkAuth
func (ac *attachmentController) putMany(c *gin.Context) {
	req, ok := BindJSON[dto.PutManyRequest](c)
	if !ok {
		return
	}
	protoFiles := make([]*gen.UploadFileDTO, len(req.Files))
	for i, f := range req.Files {
		protoFiles[i] = &gen.UploadFileDTO{
			StorageType: gen.StorageType(f.StorageType),
			FileName:    f.FileName,
			FileContent: f.FileContent,
		}
	}
	resp, err := ac.client.PutMany(c.Request.Context(), &gen.PutManyRequest{
		Files: protoFiles,
		Async: req.Async,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.PutManyResponse{IDs: resp.Ids})
}

// getOne godoc
//
//	@Summary		Получить временную ссылку на файл
//	@Description	Возвращает подписанную временную ссылку для скачивания файла по его UUID
//	@Tags			attachments
//	@Produce		json
//	@Param			X-Request-Id	header		string	true	"Уникальный идентификатор запроса"
//	@Param			id				path		string	true	"UUID файла"
//	@Param			ttl				query		int		false	"Время жизни ссылки в секундах (по умолчанию 3600)"
//	@Success		200				{object}	dto.GetOneResponse
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		404				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/attachments/{id} [get]
//	@Security		VkAuth
func (ac *attachmentController) getOne(c *gin.Context) {
	id := c.Param(ParamID)
	ttl := parseTTL(c)
	resp, err := ac.client.GetOne(c.Request.Context(), &gen.GetOneRequest{
		Id:  id,
		Ttl: ttl,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.GetOneResponse{URL: resp.Url})
}

// getMany godoc
//
//	@Summary		Получить временные ссылки на несколько файлов
//	@Description	Возвращает список подписанных временных ссылок для переданных UUID
//	@Tags			attachments
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string				true	"Уникальный идентификатор запроса"
//	@Param			body			body		dto.GetManyRequest	true	"Список UUID и TTL"
//	@Success		200				{object}	dto.GetManyResponse
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		500				{object}	dto.ErrorResponse
//	@Router			/v1/attachments [get]
//	@Security		VkAuth
func (ac *attachmentController) getMany(c *gin.Context) {
	req, ok := BindJSON[dto.GetManyRequest](c)
	if !ok {
		return
	}
	resp, err := ac.client.GetMany(c.Request.Context(), &gen.GetManyRequest{
		Ids: req.IDs,
		Ttl: req.TTL,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.GetManyResponse{URLs: resp.Urls})
}

// deleteOne godoc
//
//	@Summary		Удалить файл
//	@Description	Удаляет файл по UUID. При force=false счётчик уменьшается; физически удаляет планировщик
//	@Tags			attachments
//	@Produce		json
//	@Param			X-Request-Id	header	string	true	"Уникальный идентификатор запроса"
//	@Param			id				path	string	true	"UUID файла"
//	@Param			force			query	boolean	false	"Принудительное немедленное удаление (по умолчанию false)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/attachments/{id} [delete]
//	@Security		VkAuth
func (ac *attachmentController) deleteOne(c *gin.Context) {
	id := c.Param(ParamID)
	force, _ := ParseBoolQuery(c, queryForce)
	_, err := ac.client.DeleteOne(c.Request.Context(), &gen.DeleteOneRequest{
		Id:    id,
		Force: force,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// deleteMany godoc
//
//	@Summary		Удалить несколько файлов
//	@Description	Удаляет список файлов по UUID. При force=false физически удаляет планировщик
//	@Tags			attachments
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header	string					true	"Уникальный идентификатор запроса"
//	@Param			body			body	dto.DeleteManyRequest	true	"Список UUID и флаг force"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/attachments [delete]
//	@Security		VkAuth
func (ac *attachmentController) deleteMany(c *gin.Context) {
	req, ok := BindJSON[dto.DeleteManyRequest](c)
	if !ok {
		return
	}
	_, err := ac.client.DeleteMany(c.Request.Context(), &gen.DeleteManyRequest{
		Ids:   req.IDs,
		Force: req.Force,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// download godoc
//
//	@Summary		Скачать файл по подписанной ссылке
//	@Description	Возвращает бинарное содержимое файла. Не требует авторизации — доступ контролируется подписью
//	@Tags			attachments
//	@Produce		application/octet-stream
//	@Param			id		query	string	true	"UUID временной ссылки"
//	@Param			sign	query	string	true	"Строка подписи"
//	@Success		200
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse	"Подпись недействительна"
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/v1/file [get]
func (ac *attachmentController) download(c *gin.Context) {
	id := c.Query(ParamID)
	sign := c.Query(querySign)
	resp, err := ac.client.Download(c.Request.Context(), &gen.DownloadRequest{
		Id:   id,
		Sign: sign,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+resp.Filename)
	c.Data(http.StatusOK, resp.ContentType, resp.Content)
}

func parseTTL(c *gin.Context) int32 {
	raw := c.Query(queryTTL)
	if raw == "" {
		return defaultTTL
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return defaultTTL
	}
	return int32(v)
}

func ConfigureAttachmentController(router *gin.RouterGroup, client gen.AttachmentServiceClient) {
	ac := &attachmentController{client: client}
	router.POST(routeAttachments, ac.putOne)
	router.POST(routeAttachmentsBatch, ac.putMany)
	router.GET(routeAttachmentByID, ac.getOne)
	router.GET(routeAttachments, ac.getMany)
	router.DELETE(routeAttachmentByID, ac.deleteOne)
	router.DELETE(routeAttachments, ac.deleteMany)
	router.GET(routeAttachmentFile, ac.download)
}
