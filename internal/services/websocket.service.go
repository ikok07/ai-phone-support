package services

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketService struct {
}

func (s *WebsocketService) StartWSSession(
	w http.ResponseWriter,
	r *http.Request,
	additionalHeaders http.Header,
	timeout time.Duration,
) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from everywhere
			return true
		},
		EnableCompression: true,
		HandshakeTimeout: timeout,
	}
	return upgrader.Upgrade(w, r, additionalHeaders)
}
