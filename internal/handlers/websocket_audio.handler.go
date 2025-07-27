package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"ai-phone-support/internal/models"
	"ai-phone-support/internal/services"
	"ai-phone-support/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/maxhawkins/go-webrtcvad"
)

const (
	// End of Speech params
	END_OF_SPEECH_SILENCE_MS = 1500 // 1.5 seconds of silence to consider end of speech
	MIN_SPEECH_DURATION_MS   = 500

	// VAD params
	SPEECH_BUFFERS_COUNT  = 10
	VAD_FRAME_DURATION_MS = 20
	MIN_SPEECH_FRAMES     = MIN_SPEECH_DURATION_MS / VAD_FRAME_DURATION_MS
	SAMPLE_RATE           = 8000
	VAD_MODE              = 3 // moderate aggressiveness

	// Audio quality thresholds
	MIN_AUDIO_ENERGY      = 100
	NOISE_FLOOR_THRESHOLD = 50
	MIN_SIGNAL_TO_NOISE   = 2

	// Adaptive thresholds
	ENERGY_HISTORY        = 100
	SPEECH_CONFIDENCE_MIN = 0.7
)

var (
	vad           *webrtcvad.VAD
	speechBuffers [SPEECH_BUFFERS_COUNT]*models.SpeechBuffer
	speechSession *models.SpeechSession
	bufferMutex   sync.RWMutex

	// Adaptive noise tracking
	energyHistory      []float64
	noiseFloor         float64
	energyHistoryIndex int
	adaptiveThreshold  float64
)

var currentSpeechBufferIndex = 0

func initVAD() error {
	var err error
	vad, err = webrtcvad.New()
	if err != nil {
		return fmt.Errorf("failed to create VAD: %v", err)
	}

	if err := vad.SetMode(VAD_MODE); err != nil {
		return fmt.Errorf("failed to set VAD mode: %v", err)
	}

	for i := 0; i < SPEECH_BUFFERS_COUNT; i++ {
		speechBuffers[i] = &models.SpeechBuffer{
			ID:          i,
			Unavailable: false,
		}
	}

	speechSession = &models.SpeechSession{
		State:           models.SPEECH_STATE_IDLE,
		CurrentBufferID: 0,
	}

	energyHistory = make([]float64, ENERGY_HISTORY)
	noiseFloor = NOISE_FLOOR_THRESHOLD
	adaptiveThreshold = MIN_AUDIO_ENERGY

	return nil
}

func calculateRMSEnergy(pcm models.PCMChunk) float64 {
	if len(pcm) == 0 {
		return 0
	}

	var sum float64
	for _, sample := range pcm {
		value := float64(sample)
		sum += value * value
	}

	return math.Sqrt(sum / float64(len(pcm)))
}

func calculateZeroCrossingRate(pcm []int16) float64 {
	if len(pcm) <= 1 {
		return 0
	}

	crossings := 0
	for i := 1; i < len(pcm); i++ {
		if (pcm[i-1] >= 0 && pcm[i] < 0) || (pcm[i-1] < 0 && pcm[i] >= 0) {
			crossings++
		}
	}

	return float64(crossings) / float64(len(pcm)-1)
}

func updateAdaptiveThreshold(energy float64, isSpeechByVAD bool) {
	energyHistory[energyHistoryIndex] = energy
	energyHistoryIndex = (energyHistoryIndex + 1) % ENERGY_HISTORY

	minEnergy := math.Inf(1)
	avgEnergy := 0.00
	validSamples := 0

	for _, e := range energyHistory {
		if e > 0 {
			if e < minEnergy {
				minEnergy = e
			}
			avgEnergy += e
			validSamples++
		}
	}

	if validSamples > 0 {
		avgEnergy /= float64(validSamples)

		if minEnergy != math.Inf(1) {
			noiseFloor = 0.9*noiseFloor + 0.1*minEnergy
		}

		adaptiveThreshold = math.Max(noiseFloor*MIN_SIGNAL_TO_NOISE, MIN_AUDIO_ENERGY)
	}
}

// High-pass filter to remove low-frequency noise
func applyNoiseReduction(pcm []int16) []int16 {
	if len(pcm) < 2 {
		return pcm
	}

	filtered := make([]int16, len(pcm))
	filtered[0] = pcm[0]

	alpha := 0.95 // filter coefficient
	for i := 1; i < len(pcm); i++ {
		filtered[i] = int16(float64(pcm[i]) - alpha*float64(pcm[i-1]))
	}

	return filtered
}

func detectSpeechWithVAD(pcm models.PCMChunk) (bool, error) {
	if len(pcm) == 0 {
		return false, nil
	}

	vadResult, err := vad.Process(SAMPLE_RATE, int16ToBytes(pcm))
	if err != nil {
		return false, err
	}

	// Calculate audio features
	energy := calculateRMSEnergy(pcm)
	zcr := calculateZeroCrossingRate(pcm)

	// Update adaaptive thresholds
	updateAdaptiveThreshold(energy, vadResult)

	// Multi-criteria speech detection
	energyCheck := energy > adaptiveThreshold
	snrCheck := energy/noiseFloor > MIN_SIGNAL_TO_NOISE

	// Zero crossing rate check
	zcrCheck := zcr > 0.01 && zcr < 0.5

	// Combine all checks
	isLikelySpeech := vadResult && energyCheck && snrCheck && zcrCheck

	confidence := 0.0
	if vadResult {
		confidence += 0.4
	}
	if energyCheck {
		confidence += 0.3
	}
	if snrCheck {
		confidence += 0.2
	}
	if zcrCheck {
		confidence += 0.1
	}

	finalDecision := isLikelySpeech && confidence >= SPEECH_CONFIDENCE_MIN

	return finalDecision, nil
}

func int16ToBytes(samples []int16) []byte {
	bytes := make([]byte, len(samples)*2)
	for i, sample := range samples {
		bytes[i*2] = byte(sample & 0xFF)
		bytes[i*2+1] = byte((sample >> 8) & 0xFF)
	}
	return bytes
}

func findAvailableBuffer() (int, error) {
	bufferMutex.RLock()
	defer bufferMutex.RUnlock()

	for i, buffer := range speechBuffers {
		buffer.Mutex.RLock()
		available := !buffer.Unavailable
		buffer.Mutex.RUnlock()

		if available {
			buffer.Mutex.Lock()
			buffer.Unavailable = true
			buffer.Mutex.Unlock()
			return i, nil
		}
	}

	return -1, errors.New("no available speech buffers")
}

func processAudio(speechBuffer *models.SpeechBuffer, twilio *services.TwilioService) {
	fmt.Println("Processing audio....")

	// n8nClient := services.NewN8NService(os.Getenv("N8N_BASE_URL"))
	// options := services.N8NTriggerMainFlowOptions{
	// 	FromNumber: twilio.FromNumber,
	// 	DialNumber: nil,
	// 	Audio:      speechBuffer.Speech,
	// }
	// workflowResponse, err := n8nClient.TriggerMainWorkflow(options)
	//
	// if err == nil && len(workflowResponse) > 0 {
	// 	if err := twilio.SendAudioFromText(workflowResponse[0].Answer); err != nil {
	// 		fmt.Printf("Failed to send audio %v\n", err)
	// 	}
	// }

	time.Sleep(3 * time.Second)
	fmt.Println("Finished processing audio...")
	speechBuffer.Speech = make(models.Speech, 0)
	speechBuffer.HasSignificantSpeech = false
	speechBuffer.Unavailable = false
}

func processSpeechSession(pcm []int16, twilio *services.TwilioService) error {
	isSpeech, err := detectSpeechWithVAD(pcm)
	if err != nil {
		return errors.New("failed to run speech detection")
	}

	now := time.Now()

	switch speechSession.State {
	case models.SPEECH_STATE_IDLE:
		if isSpeech {
			bufferID, err := findAvailableBuffer()
			if err != nil {
				fmt.Println("No buffers available")
				return nil
			}

			speechSession.State = models.SPEECH_STATE_DETECTED
			speechSession.SpeechStartTime = now
			speechSession.LastSpeechTime = now
			speechSession.ConsecutiveSilentFrames = 0
			speechSession.TotalSpeechFrames = 1
			speechSession.AudioBuffer = models.Speech{pcm}
			speechSession.CurrentBufferID = bufferID

			buffer := speechBuffers[bufferID]
			buffer.Mutex.Lock()
			buffer.Speech = models.Speech{pcm}
			buffer.HasSignificantSpeech = false
			buffer.Mutex.Unlock()

			// Stop all streamed audios
			if err := twilio.StopCurrAudio(twilio.WebsocketConnection, twilio.StreamSid); err != nil {
				fmt.Printf("Error stopping audio: %v\n", err)
			}

			fmt.Println("Speech started")
		}
	case models.SPEECH_STATE_DETECTED:
		speechSession.AudioBuffer = append(speechSession.AudioBuffer, pcm)

		if isSpeech {
			speechSession.LastSpeechTime = now
			speechSession.ConsecutiveSilentFrames = 0
			speechSession.TotalSpeechFrames++
		} else {
			speechSession.ConsecutiveSilentFrames++
		}

		speechDuration := now.Sub(speechSession.SpeechStartTime)

		// Check if we should transition to waiting for end of speech
		if speechDuration > time.Duration(MIN_SPEECH_DURATION_MS)*time.Millisecond {
			speechSession.State = models.SPEECH_STATE_END_OF_SPEECH
			fmt.Println("Waiting for end of speech...")
		}

	case models.SPEECH_STATE_END_OF_SPEECH:
		speechSession.AudioBuffer = append(speechSession.AudioBuffer, pcm)

		if isSpeech {
			speechSession.LastSpeechTime = now
			speechSession.ConsecutiveSilentFrames = 0
			speechSession.TotalSpeechFrames++
		} else {
			speechSession.ConsecutiveSilentFrames++
		}

		silenceDuration := time.Duration(speechSession.ConsecutiveSilentFrames*VAD_FRAME_DURATION_MS) * time.Millisecond

		if silenceDuration >= time.Duration(END_OF_SPEECH_SILENCE_MS)*time.Millisecond {
			fmt.Println("End of speech detected")
			return endSpeechSession(twilio)
		}
	}
	return nil
}

func endSpeechSession(twilio *services.TwilioService) error {
	if speechSession.State == models.SPEECH_STATE_IDLE {
		return nil
	}

	bufferID := speechSession.CurrentBufferID
	speechDuration := time.Since(speechSession.SpeechStartTime)

	if speechSession.TotalSpeechFrames < MIN_SPEECH_FRAMES {
		fmt.Printf("Speech is too short (%d frames), ignoring\n", speechSession.TotalSpeechFrames)
		resetSpeechSession()
		return nil
	}

	currentBuffer := speechBuffers[bufferID]
	currentBuffer.Mutex.Lock()
	currentBuffer.HasSignificantSpeech = true
	currentBuffer.Mutex.Unlock()

	fmt.Printf("Processing speech: duration=%.2fs, frames=%d\n", speechDuration.Seconds(), len(speechSession.AudioBuffer))

	go func(buffer *models.SpeechBuffer) {
		defer releaseBuffer(buffer.ID)
		processAudio(buffer, twilio)
	}(currentBuffer)

	resetSpeechSession()
	return nil
}

func resetSpeechSession() {
	speechSession.State = models.SPEECH_STATE_IDLE
	speechSession.SpeechStartTime = time.Time{}
	speechSession.LastSpeechTime = time.Time{}
	speechSession.ConsecutiveSilentFrames = 0
	speechSession.TotalSpeechFrames = 0
	speechSession.AudioBuffer = nil
}

func releaseBuffer(bufferID int) {
	if bufferID < 0 || bufferID >= SPEECH_BUFFERS_COUNT {
		return
	}

	buffer := speechBuffers[bufferID]
	buffer.Mutex.Lock()
	defer buffer.Mutex.Unlock()

	buffer.Speech = nil
	buffer.HasSignificantSpeech = false
	buffer.Unavailable = false

	fmt.Printf("Released buffer %d\n", bufferID)
}

func handleStartMessage(msg []byte, twilio *services.TwilioService) error {
	fmt.Println(fmt.Sprintf("Start message %v", string(msg)))
	var startMsg models.TwilioWSStartMessage
	err := json.Unmarshal(msg, &startMsg)
	if err != nil {
		return err
	}

	twilio.FromNumber = startMsg.Start.CustomParameters["fromNumber"]
	twilio.StreamSid = startMsg.StreamSid

	greeting, err := utils.GetGreetingAudio()
	if err != nil {
		return errors.New("failed to get default greeting audio")
	}

	if err := twilio.SendAudio(twilio.WebsocketConnection, greeting); err != nil {
		return err
	}

	return nil
}

func handleMediaMessage(msg []byte, twilio *services.TwilioService) error {
	var mediaMsg models.TwilioWSMediaMessage
	err := json.Unmarshal(msg, &mediaMsg)
	if err != nil {
		return errors.New("failed to parse message")
	}

	raw, err := base64.StdEncoding.DecodeString(mediaMsg.Media.Payload)
	if err != nil {
		return errors.New("failed to convert encoded audio")
	}

	pcmAudio := utils.UlawToPcm(raw)

	return processSpeechSession(pcmAudio, twilio)
}

func handleDTMFMessage(msg []byte) error {
	var dtmfMsg models.TwilioWSDTMFMessage
	if err := json.Unmarshal(msg, &dtmfMsg); err != nil {
		return errors.New("invalid message format")
	}
	// Send the pressed digit to AI Agent
	// dtmfMsg.DTMF.Digit
	return nil
}

func WebsocketAudioHandler(c *gin.Context) {
	ws := services.WebsocketService{}
	conn, err := ws.StartWSSession(c.Writer, c.Request, nil, time.Second*30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade request to websocket!"})
		return
	}
	defer conn.Close()

	twilio := services.NewTwilioService(
		os.Getenv("TWILLIO_AUTH_KEY"),
		conn,
	)

	if err := initVAD(); err != nil {
		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Could not start Voice Active Detection"),
		)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Could not receive message"),
			)
			break
		}

		var genericMsg models.TwilioWSGenericMessage
		err = json.Unmarshal(msg, &genericMsg)
		if err != nil {
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Invalid message format"),
			)
			break
		}

		switch genericMsg.Event {
		case "start":
			err = handleStartMessage(msg, &twilio)
			if err != nil {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Invalid message format"),
				)
				break
			}
			break
		case "media":
			// fmt.Println(fmt.Sprintf("Media message %v", string(msg)))
			err = handleMediaMessage(msg, &twilio)
			if err != nil {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseUnsupportedData, err.Error()),
				)
				break
			}
			break
		case "dtmf":
			fmt.Println(fmt.Sprintf("DTMF message %v", string(msg)))
			err = handleDTMFMessage(msg)
			if err != nil {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseUnsupportedData, err.Error()),
				)
				break
			}
			break
		case "stop":
			fmt.Println(fmt.Sprintf("Stop message %v", string(msg)))
			break
		default:
			fmt.Println(fmt.Sprintf("Unhandled message %v", string(msg)))
		}
	}
}
