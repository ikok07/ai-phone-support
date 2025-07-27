package services

import (
	"encoding/base64"
	"errors"
	"os"

	"ai-phone-support/internal/constants/elevenlabs"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twilio/twilio-go/client"
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
	AuthKey             string
	WebsocketConnection *websocket.Conn
	StreamSid           string
	FromNumber          string
}

func NewTwilioService(authKey string, wsConnection *websocket.Conn) TwilioService {
	return TwilioService{
		AuthKey:             authKey,
		WebsocketConnection: wsConnection,
		StreamSid:           "",
		FromNumber:          "",
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

func (s *TwilioService) SendAudioFromText(text string) error {
	tts := NewElevenLabsService(os.Getenv("ELEVENLABS_API_KEY"))

	audio, err := tts.GenerateAudio(text, elevenlabs.ELEVENLABS_FORMAT_ULAW)
	if err != nil {
		return err
	}

	return s.SendAudio(s.WebsocketConnection, audio)
}

func (s *TwilioService) SendAudio(conn *websocket.Conn, audio []byte) error {
	encodedAudio := base64.StdEncoding.EncodeToString(audio)
	body := TwilioSendAudioBody{
		Event:     "media",
		StreamSid: s.StreamSid,
		Media: TwilioSendAudioBodyMedia{
			Payload: encodedAudio,
		},
	}

	if err := conn.WriteJSON(body); err != nil {
		return err
	}

	return nil
}

func (s *TwilioService) StopCurrAudio(conn *websocket.Conn, streamSid string) error {
	body := TwilioStopCurrAudioBody{
		Event:     "media",
		StreamSid: streamSid,
	}

	if err := conn.WriteJSON(body); err != nil {
		return errors.New("failed to stop current twilio audio")
	}

	return nil
}
