package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"os"

	"ai-phone-support/internal/constants/elevenlabs"
)

type ElevenLabsAudioChunk struct {
	Audio   *string `json:"audio"`
	IsFinal bool    `json:"isFinal"`
}

type ElevenLabsGenerateAudioBody struct {
	Text         string  `json:"text"`
	ModelId      string  `json:"model_id"`
	LanguageCode *string `json:"language_code"`
}

type ElevenLabsService struct {
	ApiKey string
}

func NewElevenLabsService(apiKey string) *ElevenLabsService {
	return &ElevenLabsService{
		ApiKey: apiKey,
	}
}

func (s *ElevenLabsService) GenerateAudio(text string, format elevenlabs.AudioFormat) ([]byte, error) {
	url, err := url2.Parse(fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", os.Getenv("ELEVENLABS_VOICE_ID")))
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("output_format", string(format))
	url.RawQuery = q.Encode()

	body := ElevenLabsGenerateAudioBody{
		Text:         text,
		ModelId:      os.Getenv("ELEVENLABS_MODEL_ID"),
		LanguageCode: nil,
	}
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("xi-api-key", s.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	resAudio, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resAudio, nil
}
