package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DownloadAudioHandler(c *gin.Context) {
	q := c.Request.URL.Query()
	filename := q.Get("filename")
	if filename == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	c.FileAttachment(fmt.Sprintf("internal/audio/%s", filename), filename)
}
