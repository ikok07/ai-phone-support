package utils

import "os"

func GetGreetingAudio() ([]byte, error) {
	return os.ReadFile("internal/audio/greeting.ulaw")
}
