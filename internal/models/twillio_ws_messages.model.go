package models

type TwilioWSGenericMessage struct {
	Event          string `json:"event"`
	SequenceNumber string `json:"sequenceNumber"`
	StreamSid      string `json:"streamSid"`
}

/* ---------- Start ---------- */

type TwilioWSStartMessage struct {
	Event          string            `json:"event"`
	SequenceNumber string            `json:"sequenceNumber"`
	Start          TwilioStartObject `json:"start"`
	StreamSid      string            `json:"streamSid"`
}

type TwilioStartObject struct {
	StreamSid        string                  `json:"streamSid"`
	AccountSid       string                  `json:"accountSid"`
	CallSid          string                  `json:"callSid"`
	Tracks           []string                `json:"tracks"`
	CustomParameters map[string]string       `json:"customParameters"`
	MediaFormat      TwilioMediaFormatObject `json:"mediaFormat"`
}

type TwilioMediaFormatObject struct {
	Encoding   string `json:"encoding"`
	SampleRate uint64 `json:"sampleRate"`
	Channels   uint8  `json:"channels"`
}

/* ---------- Media ---------- */

type TwilioWSMediaMessage struct {
	Event          string            `json:"event"`
	SequenceNumber string            `json:"sequenceNumber"`
	Media          TwilioMediaObject `json:"media"`
	StreamSid      string            `json:"streamSid"`
}

type TwilioMediaObject struct {
	Track     string `json:"track"`
	Chunk     string `json:"chunk"`
	Timestamp string `json:"timestamp"`
	Payload   string `json:"payload"`
}

/* ---------- DTMF ---------- */

type TwilioWSDTMFMessage struct {
	Event          string           `json:"event"`
	SequenceNumber string           `json:"sequenceNumber"`
	DTMF           TwilioDTMFObject `json:"media"`
	StreamSid      string           `json:"streamSid"`
}

type TwilioDTMFObject struct {
	Track string `json:"track"`
	Digit string `json:"digit"`
}
