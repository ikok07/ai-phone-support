package utils

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/go-audio/wav"
)

func WavToPcm(wavAudio []byte) ([]byte, error) {
	buffer := bytes.NewReader(wavAudio)
	decoder := wav.NewDecoder(buffer)

	if !decoder.IsValidFile() {
		return nil, errors.New("invalid WAV file")
	}

	pcmBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, errors.New("failed to decode WAV file")
	}

	var pcmData bytes.Buffer
	for _, sample := range pcmBuffer.Data {
		err := binary.Write(&pcmData, binary.LittleEndian, int16(sample))
		if err != nil {
			return nil, errors.New("failed to write PCM")
		}
	}

	return pcmData.Bytes(), nil
}
