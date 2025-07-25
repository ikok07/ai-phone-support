package routes

import (
	"ai-phone-support/internal/handlers"
	"github.com/gin-gonic/gin"
)

func callsRoutes(g *gin.RouterGroup) {
	g.POST("/receive", handlers.ReceiveCallHandler)
	g.GET("/audio", handlers.WebsocketAudioHandler)
}
