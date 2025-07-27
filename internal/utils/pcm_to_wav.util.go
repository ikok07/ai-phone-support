package utils

import (
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func PcmToWav(pcm []int16, file *os.File) error {
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
