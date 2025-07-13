package app

import (
	"context"
	"errors"
	"ggb_server/internal/app/api"
	"ggb_server/internal/config"
	"ggb_server/internal/pkg/glog"
	"ggb_server/internal/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func Start() {
	rootPath, _ := utils.FindRootPath()
	glog.InitLogger(filepath.Join(rootPath, "logs/app.log"))

	gin.SetMode(gin.DebugMode)

	e := gin.New()
	api.AddMiddleware(e)
	api.AddPath(e)

	srv := &http.Server{
		Addr:              ":" + config.Cfg.Server.Port,
		Handler:           e,
		ReadHeaderTimeout: 30 * time.Minute,
		WriteTimeout:      30 * time.Minute,
		IdleTimeout:       30 * time.Minute,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("ggb server started on :" + config.Cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second) //设置请求10min超时
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
