package http

import (
	"log"

	"github.com/gorilla/websocket"
)

// filterConnection filters initiated connection and returns true if everything
// is ok, otherwise returns false and closes the connection with a client.
func (s *server) filterConnection(clientWS *websocket.Conn, clientID uint32) bool {
	if s.checkClientIsConnected(clientID) {
		log.Printf("client %d is already connected", clientID)
		if err := clientWS.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseNormalClosure,
				"",
			)); err != nil {
			log.Println("write close:", err)
			return false
		}
		return false
	}
	return true
}

func (s *server) checkClientIsConnected(clientID uint32) bool {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.connectedClients[clientID]; ok {
		return true
	}
	// connecting new client
	s.connectedClients[clientID] = struct{}{}
	return false
}

func (s *server) disconnectClient(clientID uint32) {
	s.Lock()
	defer s.Unlock()
	delete(s.connectedClients, clientID)
}
