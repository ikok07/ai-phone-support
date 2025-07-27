package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"ai-phone-support/internal/models"
	"ai-phone-support/internal/utils"
)

type N8NService struct {
	BaseUrl string
}

type N8NTriggerMainFlowOptions struct {
	FromNumber string        `json:"fromNumber"`
	DialNumber *string       `json:"dialNumber"`
	Audio      models.Speech `json:"audio"`
}

func NewN8NService(baseUrl string) N8NService {
	return N8NService{
		BaseUrl: baseUrl,
	}
}

func (s *N8NService) TriggerMainWorkflow(options N8NTriggerMainFlowOptions) (models.N8NMainWorkflowResponse, error) {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	// Add the file if available
	if options.Audio != nil {
		if err := setAudioFormData(writer, options.Audio); err != nil {
			return nil, err
		}
	}

	// Dial number
	if options.DialNumber != nil {
		if err := writer.WriteField("dialNumber", *options.DialNumber); err != nil {
			return nil, err
		}
	}

	// Customer number
	if err := writer.WriteField("fromNumber", options.FromNumber); err != nil {
		return nil, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%v%v", s.BaseUrl, os.Getenv("N8N_MAIN_WORKFLOW_URI")), &body)
	if err != nil {
		return nil, errors.New("failed to create HTTP request")
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("workflows start request failed")
	}

	defer res.Body.Close()

	readBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("failed to read the workflow response body")
	}

	var formattedResponse models.N8NMainWorkflowResponse
	if err := json.Unmarshal(readBody, &formattedResponse); err != nil {
		fmt.Println(string(readBody))
		return nil, errors.New("failed to decode response body")
	}

	return formattedResponse, nil
}

func setAudioFormData(writer *multipart.Writer, audio models.Speech) error {
	filename := fmt.Sprintf("temp_audio_%v.wav", time.Now().Unix())
	file, err := os.Create(filename)
	if err != nil {
		return errors.New("failed to create temporary audio file")
	}
	defer file.Close()
	defer os.Remove(filename)

	var wholeSpeechPCM []int16
	for _, speechPart := range audio {
		wholeSpeechPCM = append(wholeSpeechPCM, speechPart...)
	}

	if err := utils.PcmToWav(wholeSpeechPCM, file); err != nil {
		return errors.New("failed to convert audio to wav format")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.New("failed to seek back to the beginning of the file")
	}

	part, err := writer.CreateFormFile("audio", filename)
	if err != nil {
		return errors.New("failed to create request form data field for audio")
	}

	if _, err := io.Copy(part, file); err != nil {
		return errors.New("failed to copy the wav file to the request form data")
	}

	return nil
}
