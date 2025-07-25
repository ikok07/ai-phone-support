package models

type TwillioWSGenericMessage struct {
	Event          string `json:"event"`
	SequenceNumber string `json:"sequenceNumber"`
	StreamSid      string `json:"streamSid"`
}

/* ---------- Start ---------- */

type TwillioWSStartMessage struct {
	Event          string             `json:"event"`
	SequenceNumber string             `json:"sequenceNumber"`
	Start          TwillioStartObject `json:"start"`
	StreamSid      string             `json:"streamSid"`
}

type TwillioStartObject struct {
	StreamSid        string                   `json:"streamSid"`
	AccountSid       string                   `json:"accountSid"`
	CallSid          string                   `json:"callSid"`
	Tracks           []string                 `json:"tracks"`
	CustomParameters map[string]string        `json:"customParameters"`
	MediaFormat      TwillioMediaFormatObject `json:"mediaFormat"`
}

type TwillioMediaFormatObject struct {
	Encoding   string `json:"encoding"`
	SampleRate uint64 `json:"sampleRate"`
	Channels   uint8  `json:"channels"`
}

/* ---------- Media ---------- */

type TwillioWSMediaMessage struct {
	Event          string             `json:"event"`
	SequenceNumber string             `json:"sequenceNumber"`
	Media          TwillioMediaObject `json:"media"`
	StreamSid      string             `json:"streamSid"`
}

type TwillioMediaObject struct {
	Track string `json:"track"`
	Chunk string `json:"chunk"`
	Timestamp string `json:"timestamp"`
	Payload string `json:"payload"`
}