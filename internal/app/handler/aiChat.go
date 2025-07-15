package handler

import (
	"fmt"
	"ggb_server/internal/app/model"
	"ggb_server/internal/app/schema"
	"ggb_server/internal/config"
	"ggb_server/internal/pkg/workflow"
	"ggb_server/internal/repository"
	"ggb_server/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type AiChat struct{}

func (a AiChat) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
	}
	fileExt := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExt)
	rootPath, _ := utils.FindRootPath()
	savePath := filepath.Join(rootPath, config.Cfg.Static.Path, newFileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
	}
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	c.JSON(200, gin.H{
		"imageUrl": fmt.Sprintf("%s://%s/%s/%s", protocol, c.Request.Host, config.Cfg.Static.Path, newFileName),
	})
}

func (a AiChat) Chat(c *gin.Context) {
	var (
		chatRequest schema.ChatRequest
	)

	if err := c.ShouldBind(&chatRequest); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	message, err := insertMessage(GetDB(c), chatRequest)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err,
		})
	}

	flusher := c.Writer.(http.Flusher)
	w := c.Writer
	c.Writer.WriteHeader(http.StatusOK)
	processor := workflow.NewProcess(chatRequest.Message, chatRequest.ImageUrl, flusher, w, c.Request.Context())
	err = processor.StartProcess(GetDB(c), message)
	if err != nil {
		log.Println("[error] occurred when processing: ", err)
	}
}

func insertMessage(db *gorm.DB, chatRequest schema.ChatRequest) (*model.Message, error) {
	message := &model.Message{
		ParentID:  chatRequest.ParentId,
		SessionID: chatRequest.SessionId,
		UserID:    utils.GenerateRandomString(36),
		Message:   chatRequest.Message,
		Identity:  0,
	}
	messageRepo := repository.NewMessageRepository()
	if err := messageRepo.Create(db, message); err != nil {
		return nil, err
	}
	if chatRequest.ImageUrl != "" {
		resource := model.Resource{
			SessionID: chatRequest.SessionId,
			MessageID: message.ID,
			Type:      1,
			URL:       utils.ProcessUrl(chatRequest.ImageUrl, config.Cfg.Static.Path),
		}
		return message, repository.NewResourceRepository().Create(db, &resource)
	}
	return message, nil
}
