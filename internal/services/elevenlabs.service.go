package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	url2 "net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ElevenLabsAudioChunk struct {
	Audio   *string `json:"audio"`
	IsFinal bool    `json:"isFinal"`
}

type ElevenLabsService struct {
	ApiKey string
}

func NewElevenLabsService(apiKey string) *ElevenLabsService {
	return &ElevenLabsService{
		ApiKey: apiKey,
	}
}

func (s *ElevenLabsService) GenerateAudioStream() (*websocket.Conn, error) {
	url, err := url2.Parse(fmt.Sprintf("wss://api.elevenlabs.io/v1/text-to-speech/%s/stream-input", os.Getenv("VOICE_ID")))
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("model_id", os.Getenv("MODEL_ID"))
	q.Set("language_code", "bg")
	url.RawQuery = q.Encode()

	headers := http.Header{}
	headers.Set("xi-api-key", s.ApiKey)

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), headers)
	if err != nil {
		return nil, err
	}

	err = conn.WriteJSON(gin.H{
		"text": " ",
	})

	if err != nil {
		return nil, err;
	}

	return conn, nil
}

func (s *ElevenLabsService) SendText(conn *websocket.Conn, text string) error {
	return conn.WriteJSON(gin.H{text: text})
}

func (s *ElevenLabsService) EndAudioStream(conn *websocket.Conn) error {
	return conn.WriteJSON(gin.H{"text": ""})
}

// ReceiveAudio - returns base64 encoded audio chunk.
func (s *ElevenLabsService) ReceiveAudio(conn *websocket.Conn) (*string, error) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		var chunk ElevenLabsAudioChunk
		if err := json.Unmarshal(msg, &chunk); err != nil {
			return nil, errors.New(fmt.Sprintf("failed to decode ElevenLabs chunk! %v", err))
		}

		if chunk.IsFinal {
			return nil, nil
		}

		if chunk.Audio == nil {
			return nil, errors.New("audio not final but is null")
		}

		return chunk.Audio, nil
	}
}
