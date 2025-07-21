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

func (a AiChat) Conversation(c *gin.Context) {
	var req schema.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 从JWT token中获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户未认证",
		})
		return
	}

	db := GetDB(c)
	sessionRepo := repository.NewSessionRepository()

	// 创建新的session
	session := &model.Session{
		Title:            req.Title,
		UserID:           userID,
		MessageCount:     0,
		FreeMessageCount: 100, // 默认免费消息额度
		IsDel:            0,
	}

	err = sessionRepo.Create(db, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建对话失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": schema.CreateConversationResponse{
			ID:    session.ID,
			Title: session.Title,
		},
	})
}

func getUserIDFromToken(c *gin.Context) (string, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return "", fmt.Errorf("未提供认证token")
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 解析token
	claims, err := utils.ParseToken(token)
	if err != nil {
		return "", err
	}

	// 检查token是否过期
	if claims.ExpiresAt < time.Now().Unix() {
		return "", fmt.Errorf("token已过期")
	}

	// 从token中获取用户信息
	db := GetDB(c)
	userRepo := repository.NewUserRepository()
	user, err := userRepo.GetById(db, int64(claims.UserID))
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	// 使用用户名作为UserID，因为Session模型中的UserID是string类型
	return user.Username, nil
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
