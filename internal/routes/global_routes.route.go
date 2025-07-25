package routes

import (
	"github.com/gin-gonic/gin"
)

func GlobalRoutes(e *gin.Engine) {
	// Calls routes
	callsGroup := e.Group("/calls")
	callsRoutes(callsGroup)
}
