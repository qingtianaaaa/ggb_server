package api

import (
	"ggb_server/internal/app/handler"
	"ggb_server/internal/middleware"
	"github.com/gin-gonic/gin"
	"os"
)

var (
	AIChatAPI handler.AiChat
)

func AddPath(e *gin.Engine) {
	err := os.MkdirAll("static/upload", os.ModePerm)
	if err != nil {
		return
	}
	e.Static("/static/upload", "./static/upload")

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
