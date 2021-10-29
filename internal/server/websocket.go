package server

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
	"test.task/backend/proxy/internal/model"
)

func (s *server) mustGetServerConn() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: s.backendAddr, Path: "/connect"}
	serverWS, _, err := s.dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial to a server:", err)
	}
	return serverWS
}

func writeToConn(conn *websocket.Conn, connType string, mt int, message []byte) error {
	if err := conn.WriteMessage(mt, message); err != nil {
		log.Printf("write to %s: %v", connType, err)
		return err
	}
	return nil
}

func writeOtherErrorToClient(clientWS *websocket.Conn, ID uint32, customErr error) {
	log.Printf("error ID %d: %v", ID, customErr)
	res := proxy.OrderResponse{
		ID:   ID,
		Code: uint16(model.ResultCodeOther),
	}
	writeToConn(clientWS, "client", websocket.TextMessage, proxy.EncodeOrderResponse(res))
}
