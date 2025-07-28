package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	elevenlabs2 "ai-phone-support/internal/constants/elevenlabs"
	"ai-phone-support/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/twiml"
)

func getErrorXML(twilio *services.TwilioService) string {
	return twilio.ListenForInput(
		"dtmf speech",
		fmt.Sprintf("%s/calls/audio", os.Getenv("BASE_URL_TUNNEL_DEV")),
		[]twiml.Element{
			twiml.VoiceParameter{Name: "fromNumber", Value: twilio.FromNumber},
		},
		[]twiml.Element{
			twiml.VoicePlay{
				Url:  fmt.Sprintf("%s/calls/audio?filename=error.ulaw", os.Getenv("BASE_URL_TUNNEL_DEV")),
				Loop: "1",
			},
		},
	)
}

func processAudio(transcription string, twilio *services.TwilioService) {
	fmt.Println("Processing audio....")

	n8nClient := services.NewN8NService(os.Getenv("N8N_BASE_URL"))
	options := services.N8NTriggerMainFlowOptions{
		FromNumber:    twilio.FromNumber,
		DialNumber:    nil,
		Transcription: &transcription,
	}
	workflowResponse, err := n8nClient.TriggerMainWorkflow(options)
	if err != nil || len(workflowResponse) == 0 {
		fmt.Printf("Workflow failed! %v", err)
		if err := twilio.UpdateCall(getErrorXML(twilio)); err != nil {
			fmt.Printf("Failed to update call: %v", err)
		}
		return
	}

	elevenlabs := services.ElevenLabsService{
		ApiKey: os.Getenv("ELEVENLABS_API_KEY"),
	}

	audio, err := elevenlabs.GenerateAudio(workflowResponse[0].Answer, elevenlabs2.ELEVENLABS_FORMAT_ULAW)
	if err != nil {
		fmt.Printf("Failed to generate audio! %v", err)
		if err := twilio.UpdateCall(getErrorXML(twilio)); err != nil {
			fmt.Printf("Failed to update call: %v", err)
		}
		return
	}

	audioFilename := fmt.Sprintf("temp-audio-%v.ulaw", time.Now().Unix())
	err = os.WriteFile(fmt.Sprintf("internal/audio/temp/%v", audioFilename), audio, 0644)
	if err != nil {
		fmt.Printf("Failed to create audio file! %v", err)
		if err := twilio.UpdateCall(getErrorXML(twilio)); err != nil {
			fmt.Printf("Failed to update call: %v", err)
		}
		return
	}

	fmt.Println("Finished processing audio...")

	xml := twilio.ListenForInput(
		"dtmf speech",
		fmt.Sprintf("%s/calls/audio", os.Getenv("BASE_URL_TUNNEL_DEV")),
		[]twiml.Element{
			twiml.VoiceParameter{Name: "fromNumber", Value: twilio.FromNumber},
		},
		[]twiml.Element{
			twiml.VoicePlay{
				Url:  fmt.Sprintf("%s/calls/audio?filename=temp/%s", os.Getenv("BASE_URL_TUNNEL_DEV"), audioFilename),
				Loop: "1",
			},
		},
	)
	fmt.Println(xml)
	if err := twilio.UpdateCall(xml); err != nil {
		fmt.Printf("Failed to update call: %v", err)
	}
}

func AudioHandler(c *gin.Context) {
	twillio := services.TwilioService{
		AuthKey: os.Getenv("TWILLIO_AUTH_KEY"),
	}
	if !twillio.ValidateRequest(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	params := c.Request.PostForm

	twillio.CallSid = params.Get("CallSid")
	twillio.FromNumber = params.Get("From")

	go processAudio(params.Get("SpeechResult"), &twillio)

	voicePlay := twiml.VoicePlay{
		Url:  "https://flw3nxfetvtbrrcydjcu5zmtde.srv.us/calls/audio?filename=wait_1.ulaw",
		Loop: "1",
	}

	voicePause := twiml.VoicePause{
		Length: "30",
	}

	xml, _ := twiml.Voice([]twiml.Element{
		voicePlay,
		voicePause,
	})

	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, xml)
}
