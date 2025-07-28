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

	"ai-phone-support/internal/models"
)

type N8NService struct {
	BaseUrl string
}

type N8NTriggerMainFlowOptions struct {
	FromNumber    string  `json:"fromNumber"`
	Transcription *string `json:"transcription"`
}

func NewN8NService(baseUrl string) N8NService {
	return N8NService{
		BaseUrl: baseUrl,
	}
}

func (s *N8NService) TriggerMainWorkflow(options N8NTriggerMainFlowOptions) (models.N8NMainWorkflowResponse, error) {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	// Add the transcription if available
	if options.Transcription != nil {
		if err := writer.WriteField("transcription", *options.Transcription); err != nil {
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
