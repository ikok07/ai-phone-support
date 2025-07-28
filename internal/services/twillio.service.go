package services

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"
)

type TwilioSendAudioBody struct {
	Event     string                   `json:"event"`
	StreamSid string                   `json:"streamSid"`
	Media     TwilioSendAudioBodyMedia `json:"media"`
}

type TwilioSendAudioBodyMedia struct {
	Payload string `json:"payload"`
}

type TwilioStopCurrAudioBody struct {
	Event     string `json:"event"`
	StreamSid string `json:"streamSid"`
}

type TwilioService struct {
	AuthKey    string
	CallSid    string
	FromNumber string
}

func NewTwilioService(authKey string) TwilioService {
	return TwilioService{
		AuthKey:    authKey,
		CallSid:    "",
		FromNumber: "",
	}
}

func (s *TwilioService) ValidateRequest(c *gin.Context) bool {
	requestValidator := client.NewRequestValidator(s.AuthKey)

	url := "https://" + c.Request.Host + c.Request.URL.RequestURI()

	if err := c.Request.ParseForm(); err != nil {
		return false
	}

	params := make(map[string]string)
	if c.Request.Method == "POST" {
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	} else {
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	signature := c.GetHeader("X-Twilio-Signature")

	return requestValidator.Validate(url, params, signature)
}

func (s *TwilioService) UpdateCall(xml string) error {
	var twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILLIO_API_KEY_SID"),
		Password: os.Getenv("TWILLIO_API_KEY_SECRET"),
	})

	params := openapi.UpdateCallParams{}
	params.SetTwiml(xml)

	if _, err := twilioClient.Api.UpdateCall(s.CallSid, &params); err != nil {
		return err
	}
	return nil
}

func (s *TwilioService) ListenForInput(inputType string, url string, gatherInnerElements []twiml.Element, beforeGatherElements []twiml.Element) string {
	voiceGather := twiml.VoiceGather{
		Action:              url,
		ActionOnEmptyResult: "false",
		Input:               inputType,
		Language:            "bg",
		Method:              "POST",
		SpeechModel:         "deepgram_nova-2",
		SpeechTimeout:       "3", // seconds
		Timeout:             "3", // seconds
		InnerElements:       gatherInnerElements,
	}

	elements := []twiml.Element{}

	elements = append(elements, beforeGatherElements...)
	elements = append(elements, voiceGather)

	xml, _ := twiml.Voice(elements)

	return xml
}
