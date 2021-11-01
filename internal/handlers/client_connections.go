package handlers

import (
	"log"

	"github.com/gorilla/websocket"
)

// filterConnection filters initiated connection and returns true if everything
// is ok, otherwise returns false and closes the connection with a client.
func (p *ProxyHandler) filterConnection(clientWS *websocket.Conn, clientID uint32) bool {
	if p.checkClientIsConnected(clientID) {
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

func (p *ProxyHandler) checkClientIsConnected(clientID uint32) bool {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.connectedClients[clientID]; ok {
		return true
	}
	// connecting new client
	p.connectedClients[clientID] = struct{}{}
	return false
}

func (p *ProxyHandler) disconnectClient(clientID uint32) {
	p.Lock()
	defer p.Unlock()
	delete(p.connectedClients, clientID)
}
