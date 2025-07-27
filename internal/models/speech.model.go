package models

import (
	"sync"
	"time"
)

type PCMChunk []int16
type Speech []PCMChunk

type SpeechBuffer struct {
	ID                   int
	Speech               Speech
	Unavailable          bool
	HasSignificantSpeech bool
	Mutex                sync.RWMutex
}

type SpeechSession struct {
	State                   SpeechState
	SpeechStartTime         time.Time
	LastSpeechTime          time.Time
	ConsecutiveSilentFrames int
	TotalSpeechFrames       int
	AudioBuffer             Speech
	CurrentBufferID         int
}

type SpeechState int

const (
	SPEECH_STATE_IDLE          SpeechState = 0
	SPEECH_STATE_DETECTED      SpeechState = 1
	SPEECH_STATE_END_OF_SPEECH SpeechState = 2
)
