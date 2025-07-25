package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"ai-phone-support/internal/models"
	"ai-phone-support/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gorilla/websocket"
	"github.com/zaf/g711"
)

const (
	SILENCE_THRESHOLD            = 1000
	AUDIO_PROCESS_PERIOD_SECONDS = 1
	MIN_SPEECH_CHUNKS            = 20
)

type PCMChunk []int16
type Speech []PCMChunk

type SpeechBuffer struct {
	speech               Speech
	unavailable          bool
	hasSignificantSpeech bool
}

func convertToPCM(base64Audio string) ([]int16, error) {
	raw, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil {
		return nil, err
	}

	pcm := make([]int16, len(raw))
	for index, data := range raw {
		pcm[index] = g711.DecodeUlawFrame(data)
	}

	return pcm, nil
}

func checkIfSilence(pcm []int16) bool {
	var max int16
	for _, value := range pcm {
		abs := value
		if abs < 0 {
			abs = -abs
		}
		if abs > max {
			max = abs
		}
	}
	return max < SILENCE_THRESHOLD
}

// NOTE: The file should be opened before calling this function
func pcmToWav(pcm []int16, file *os.File) error {
	encoder := wav.NewEncoder(file, 8000, 16, 1, 1)
	defer encoder.Close()
	buffer := audio.IntBuffer{
		Format: &audio.Format{NumChannels: 1, SampleRate: 8000},
		Data:   make([]int, len(pcm)),
	}
	for i, value := range pcm {
		buffer.Data[i] = int(value)
	}
	return encoder.Write(&buffer)
}

func processAudio(speechBuffer *SpeechBuffer) {
	fmt.Println("Processing audio....")
	file, err := os.Create("test.wav")
	if err != nil {
		return
	}
	defer file.Close()

	var wholeSpeechPCM []int16
	for _, speechPart := range speechBuffer.speech {
		wholeSpeechPCM = append(wholeSpeechPCM, speechPart...)
	}

	if len(wholeSpeechPCM) == 0 {
		return
	}

	if err := pcmToWav(wholeSpeechPCM, file); err != nil {
		return
	}

	time.Sleep(time.Second * 3)
	fmt.Println("Finished processing audio...")
	speechBuffer.speech = make(Speech, 0)
	speechBuffer.hasSignificantSpeech = false
	speechBuffer.unavailable = false
}

func findAvailableBuffer(speechBuffers []SpeechBuffer, currentIndex uint8) (uint8, bool) {
	for i := 1; i < len(speechBuffers); i++ {
		nextIndex := (int(currentIndex) + i) % len(speechBuffers)
		if !speechBuffers[nextIndex].unavailable {
			return uint8(nextIndex), true
		}
	}
	return currentIndex, false
}

func WebsocketAudioHandler(c *gin.Context) {
	ws := services.WebsocketService{}
	conn, err := ws.StartWSSession(c.Writer, c.Request, nil, time.Second*30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade request to websocket!"})
		return
	}
	defer conn.Close()

	var lastSpeech time.Time
	var currentSpeechBufferIndex uint8 = 0
	speechBuffers := make([]SpeechBuffer, 10)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Could not receive message"),
			)
			break
		}

		var genericMsg models.TwillioWSGenericMessage
		err = json.Unmarshal(msg, &genericMsg)
		if err != nil {
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Invalid message format"),
			)
			break
		}

		// var fromNumber string

		if genericMsg.Event == "start" {
			fmt.Println(fmt.Sprintf("Start message %v", string(msg)))
			var startMsg models.TwillioWSStartMessage
			err = json.Unmarshal(msg, &startMsg)
			if err != nil {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Invalid message format"),
				)
				break
			}

		} else if genericMsg.Event == "media" {
			// fmt.Println(fmt.Sprintf("Media message %v", string(msg)))
			var mediaMsg models.TwillioWSMediaMessage
			err = json.Unmarshal(msg, &mediaMsg)
			if err != nil {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Invalid message format"),
				)
				break
			}

			pcmAudio, err := convertToPCM(mediaMsg.Media.Payload)
			if err != nil {
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Failed to convert encoded audio"),
				)
				break
			}

			speechBuffers[currentSpeechBufferIndex].speech = append(speechBuffers[currentSpeechBufferIndex].speech, pcmAudio)

			if !checkIfSilence(pcmAudio) {
				lastSpeech = time.Now()
				speechBuffers[currentSpeechBufferIndex].hasSignificantSpeech = true
			}

			shouldProcess := !lastSpeech.IsZero() &&
				time.Since(lastSpeech) > time.Second*AUDIO_PROCESS_PERIOD_SECONDS &&
				len(speechBuffers[currentSpeechBufferIndex].speech) > MIN_SPEECH_CHUNKS &&
				speechBuffers[currentSpeechBufferIndex].hasSignificantSpeech

			if shouldProcess {
				currIndex := currentSpeechBufferIndex
				speechBuffers[currIndex].unavailable = true
				if nextIndex, ok := findAvailableBuffer(speechBuffers, currIndex); ok {
					currentSpeechBufferIndex = nextIndex
				} else {
					_ = conn.WriteMessage(
						websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Audio buffer not available"),
					)
					return
				}

				go processAudio(&speechBuffers[currIndex])
			}
		} else if genericMsg.Event == "stop" {
			fmt.Println(fmt.Sprintf("Stop message %v", string(msg)))
		}
	}
}
