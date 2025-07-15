package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDB(c *gin.Context) *gorm.DB {
	return c.MustGet("db").(*gorm.DB)
}
