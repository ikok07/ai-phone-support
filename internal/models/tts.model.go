package models

import "github.com/gorilla/websocket"

type TTSService interface {
	GenerateAudioStream() (*websocket.Conn, error)
	SendText(conn *websocket.Conn, text string) error
	EndAudioStream(conn *websocket.Conn) error
	ReceiveAudio(conn *websocket.Conn) (*string, error)
}

type TTSServiceGenerateAudioResponse struct {
}
