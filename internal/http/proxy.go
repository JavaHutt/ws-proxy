package http

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
)

func (s *server) proxyHandler(w http.ResponseWriter, r *http.Request) {
	clientWS, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade client request:", err)
		return
	}

	// reading message first time not in a loop because firstly
	// we need to get client id which is inside binary message
	mt, message, err := clientWS.ReadMessage()
	if err != nil {
		return
	}
	req := proxy.DecodeOrderRequest(message)
	clientID := req.ClientID

	// checking initial connection
	filterPassed := s.filterConnection(clientWS, clientID)
	if !filterPassed {
		return
	}

	serverWS := s.mustGetServerConn()
	// firstProcess runs once when connection had been established
	s.firstProcess(clientWS, serverWS, req, mt, message)

	// start listening from server and repeat message directly to client
	go s.serverToClient(serverWS, clientWS)
	// process client message and pass it to server if everything is ok
	s.clientToServer(clientWS, serverWS, clientID)
}

func (s *server) clientToServer(clientWS, serverWS *websocket.Conn, clientID uint32) {
	defer clientWS.Close()
	for {
		mt, message, err := clientWS.ReadMessage()
		if err != nil {
			s.disconnectClient(clientID)
			break
		}
		req := proxy.DecodeOrderRequest(message)
		log.Printf("recv from client: %v", req)
		id := req.ID

		translatedOrder, err := s.adapter.TranslateOrder(req)
		if err != nil {
			s.writeErrorToClient(clientWS, id, err)
			continue
		}
		if err := s.svc.ProcessOrder(translatedOrder); err != nil {
			s.writeErrorToClient(clientWS, id, err)
			continue
		}

		if err = writeToConn(serverWS, "server", mt, message); err != nil {
			continue
		}

		log.Printf("sent to server: %v", req)
	}
}

func (s *server) serverToClient(serverWS, clientWS *websocket.Conn) {
	defer serverWS.Close()
	for {
		mt, messsage, err := serverWS.ReadMessage()
		if err != nil {
			log.Printf("recv error: %+v", err)
			return
		}
		res := proxy.DecodeOrderResponse(messsage)

		if err = writeToConn(clientWS, "client", mt, messsage); err != nil {
			continue
		}

		log.Printf("recv from server and sent to client: %v", res)
	}
}

func (s *server) firstProcess(
	clientWS, serverWS *websocket.Conn,
	req proxy.OrderRequest,
	mt int,
	message []byte,
) {
	id := req.ID
	translatedOrder, err := s.adapter.TranslateOrder(req)
	if err != nil {
		// the task description didn't specify the way to respond to invalid
		// requests, so I've decided to send back "Other" result code
		s.writeErrorToClient(clientWS, id, err)
		return
	}
	if err = s.svc.ProcessOrder(translatedOrder); err != nil {
		s.writeErrorToClient(clientWS, id, err)
		return
	}
	// first-time write to server after establishing connection with client
	writeToConn(serverWS, "server", mt, message)
}
