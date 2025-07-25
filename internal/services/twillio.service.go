package services

import (
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/client"
)

type TwillioService struct {
	AuthKey string
}

func (s *TwillioService) ValidateRequest(c *gin.Context) bool {
	requestValidator := client.NewRequestValidator(s.AuthKey)

	url := "https://" + c.Request.Host + c.Request.URL.RequestURI()

	paramsRaw := c.Request.URL.Query()
	params := make(map[string]string)
	for key, value := range paramsRaw {
		params[key] = value[0]
	}

	signature := c.GetHeader("X-Twilio-Signature")

	return requestValidator.Validate(url, params, signature)
}
