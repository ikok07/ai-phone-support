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

	voiceStream := twiml.VoiceStream{
		Name:          "Test",
		Url:           "wss://flw3nxfetvtbrrcydjcu5zmtde.srv.us/calls/audio",
		InnerElements: []twiml.Element{fromNumberParam},
	}

	xml, err := twiml.Voice([]twiml.Element{
		twiml.VoiceConnect{InnerElements: []twiml.Element{voiceStream}},
		twiml.VoiceSay{Message: "Hello"},
		twiml.VoicePause{Length: "3"},
		twiml.VoiceSay{Message: "This is a test"},
		twiml.VoicePause{Length: "3"},
		twiml.VoiceSay{Message: "Goodbye"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate new voice stream!"})
		return
	}

	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, xml)
}
