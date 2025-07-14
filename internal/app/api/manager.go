package api

import (
	"ggb_server/internal/app/handler"
	"ggb_server/internal/config"
	"ggb_server/internal/middleware"
	"ggb_server/internal/pkg/database"
	"ggb_server/internal/utils"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
)

var (
	AIChatAPI handler.AiChat
)

func AddPath(e *gin.Engine) {
	rootPath, _ := utils.FindRootPath()
	uploadDir := filepath.Join(rootPath, config.Cfg.Static.Path)
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return
	}
	e.Static(config.Cfg.Static.Path, uploadDir)

	e.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r := e.Group("/api")
	aiChatModule(r)
	userModule(r)
}

func AddMiddleware(e *gin.Engine) {
	e.Use(middleware.CORSMiddleware())
	e.Use(middleware.Logger())
	e.Use(middleware.Recovery())
	e.Use(middleware.DBMiddleware(database.GetDB()))
	//e.Use(middleware.JWTAuthMiddleware())
}

func aiChatModule(r *gin.RouterGroup) {
	r.POST("/upload", AIChatAPI.Upload)
	r.POST("/v2/chat", AIChatAPI.Chat)
}

func userModule(r *gin.RouterGroup) {
	r = r.Group("/user")
	{

	}
}
