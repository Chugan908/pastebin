package initialize

import (
	"github.com/gin-gonic/gin"
	"pastebin/internal/controllers"
)

func SetupRoutes(handler *controllers.Handler, server *gin.Engine) {
	server.POST("/read/", handler.OpenText())
	server.POST("/create_text", handler.CreateText())
}
