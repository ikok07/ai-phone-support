package elevenlabs

type AudioFormat string

const (
	ELEVENLABS_FORMAT_DEFAULT AudioFormat = "mp3_44100_128"
	ELEVENLABS_FORMAT_ULAW    AudioFormat = "ulaw_8000"
)
