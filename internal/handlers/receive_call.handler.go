package handlers

import (
	"fmt"
	"net/http"
	"os"

	"ai-phone-support/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/twiml"
)

func ReceiveCallHandler(c *gin.Context) {
	twillio := services.TwilioService{
		AuthKey: os.Getenv("TWILLIO_AUTH_KEY"),
	}
	if !twillio.ValidateRequest(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Save this to a db
	fromNumber := c.Request.PostForm["From"][0]
	if fromNumber != "client:Anonymous" {
		// Save to db...
	}
	fmt.Println(fromNumber)

	fromNumberParam := twiml.VoiceParameter{
		Name:  "fromNumber",
		Value: fromNumber,
	}

	xml := twillio.ListenForInput(
		"dtmf speech",
		fmt.Sprintf("%s/calls/audio", os.Getenv("BASE_URL_TUNNEL_DEV")),
		[]twiml.Element{
			fromNumberParam,
		},
		[]twiml.Element{
			twiml.VoicePlay{
				Loop: "1",
				Url:  fmt.Sprintf("%s/calls/audio?filename=greeting.ulaw", os.Getenv("BASE_URL_TUNNEL_DEV")),
			},
		},
	)

	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, xml)
}
