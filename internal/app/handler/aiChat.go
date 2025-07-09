package handler

import (
	"fmt"
	"ggb_server/internal/app/schema"
	"ggb_server/internal/pkg/workflow"
	"ggb_server/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	currentDir, _ := os.Getwd()
	savePath := filepath.Join(currentDir, "static/upload", newFileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
	}
	imageUrl := fmt.Sprintf("http://localhost:8080/static/upload/%s", newFileName)
	c.JSON(200, gin.H{
		"imageUrl": imageUrl,
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

	flusher := c.Writer.(http.Flusher)
	w := c.Writer
	c.Writer.WriteHeader(http.StatusOK)
	processor := workflow.NewProcess(chatRequest.Message, flusher, w)
	rootPath, _ := utils.FindRootPath()
	err := processor.StartProcess(strings.Replace(chatRequest.ImageUrl, "http://localhost:8080", rootPath, 1))
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
	}
}
