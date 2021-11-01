package handlers

import (
	"log"

	"github.com/gorilla/websocket"
)

// filterConnection filters initiated connection and returns true if everything
// is ok, otherwise returns false and closes the connection with a client.
func (p *ProxyHandler) filterConnection(clientWS *websocket.Conn, clientID uint32) bool {
	if !p.clientsSvc.TryConnectClient(clientID) {
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
