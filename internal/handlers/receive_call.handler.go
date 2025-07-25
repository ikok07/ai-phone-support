package handlers

import (
	"net/http"
	// "os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/twiml"
	// "ai-phone-support/internal/services"
)

func ReceiveCallHandler(c *gin.Context) {
	// twillio := services.TwillioService{
	// 	AuthKey: os.Getenv("TWILLIO_SECRET_KEY"),
	// }
	// if !twillio.ValidateRequest(c) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	// params := c.Request.URL.Query()

	// Save this to a db
	// fromNumber := params["From"][0]
	// if fromNumber != "client:Anonymous" {
	// 	// Save to db...
	// }
	// fmt.Println(fromNumber)

	voiceStream := twiml.VoiceStream{
		Name: "Test",
		Url:  "wss://flw3nxfetvtbrrcydjcu5zmtde.srv.us/calls/audio",
		// OptionalAttributes: map[string]string{fromNumber: fromNumber},
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
