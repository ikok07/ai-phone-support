package handlers

import (
	"fmt"
	"net/http"
	"os"

	"ai-phone-support/internal/db"
	"ai-phone-support/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/twiml"
	"gorm.io/gorm/clause"
)

func ReceiveCallHandler(c *gin.Context) {
	twillio := services.TwilioService{
		AuthKey: os.Getenv("TWILLIO_AUTH_KEY"),
	}
	if !twillio.ValidateRequest(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Create new customer if not available
	fromNumber := c.Request.PostForm["From"][0]
	if fromNumber == "client:Anonymous" {
		fromNumber = os.Getenv("TWILLIO_DUMMY_PHONE")
	}

	customer := db.Customer{
		PhoneNumber: fromNumber,
	}
	err := db.Insert(customer, db.DB.Clauses(clause.OnConflict{DoNothing: true}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

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
