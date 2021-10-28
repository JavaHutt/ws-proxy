package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
)

// Server will perform operations over http.
type Server interface {
	// Open will setup a tcp listener and serve the http requests.
	Open() error

	// Close will close the socket if it's open.
	Close(ctx context.Context) error
}

// Server represents an HTTP server.
type server struct {
	sync.Mutex
	addr             string
	backendAddr      string
	connectedClients map[uint32]struct{}
	serv             *http.Server
	upgrader         websocket.Upgrader
	dialer           *websocket.Dialer
}

func NewServer(addr, backendAddr string) Server {
	return &server{
		addr:             addr,
		backendAddr:      backendAddr,
		connectedClients: make(map[uint32]struct{}),
		upgrader:         websocket.Upgrader{},
		dialer:           websocket.DefaultDialer,
	}
}

// Open will setup a tcp listener and serve the http requests.
func (s *server) Open() error {
	s.serv = &http.Server{
		Addr: s.addr,
	}
	log.Printf("Waiting for connections on %s/", s.addr)
	http.HandleFunc("/", s.proxyHandler)
	return s.serv.ListenAndServe()
}

// Close will close the socket if it's open.
func (s *server) Close(ctx context.Context) error {
	if s.serv != nil {
		if err := s.serv.Shutdown(ctx); err != nil {
			return err
		}
		s.serv = nil
	}
	return nil
}

func (s *server) proxyHandler(w http.ResponseWriter, r *http.Request) {
	clientWS, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade client request:", err)
		return
	}

	u := url.URL{Scheme: "ws", Host: s.backendAddr, Path: "/connect"}
	serverWS, _, err := s.dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial to a server:", err)
	}

	done := make(chan struct{})
	go s.serverToClient(serverWS, clientWS, done)
	go s.clientToServer(clientWS, serverWS)
}

func (s *server) clientToServer(clientWS, serverWS *websocket.Conn) {
	defer clientWS.Close()
	for {
		mt, message, err := clientWS.ReadMessage()
		if err != nil {
			break
		}
		req := proxy.DecodeOrderRequest(message)
		log.Printf("recv from client: %v", req)
		// TODO No idea how to store information about connected user...
		// clientID := req.ClientID
		// if s.checkClientIsConnected(clientID) {
		// 	log.Printf("client %d is already connected", clientID)
		// 	break
		// }

		fmt.Println("process proxy req")

		res := proxy.OrderResponse{
			ID:   req.ID,
			Code: 0,
		}
		if err = serverWS.WriteMessage(mt, message); err != nil {
			log.Println("write to server:", err)
			continue
		}

		log.Printf("sent: %v", res)
	}
}

func (s *server) serverToClient(serverWS, clientWS *websocket.Conn, done chan struct{}) {
	defer serverWS.Close()
	for {
		mt, messsage, err := serverWS.ReadMessage()
		if err != nil {
			log.Printf("recv error: %+v", err)
			return
		}
		decoded := proxy.DecodeOrderResponse(messsage)

		if err = clientWS.WriteMessage(mt, messsage); err != nil {
			log.Println("write to client:", err)
			continue
		}

		log.Printf("recv from server and sent to client: %v", decoded)
	}
}

func (s *server) checkClientIsConnected(clientID uint32) bool {
	if _, ok := s.connectedClients[clientID]; ok {
		return true
	}
	s.Lock()
	defer s.Unlock()
	s.connectedClients[clientID] = struct{}{}
	return false
}

func (s *server) disconnectClient(clientID uint32) {
	s.Lock()
	defer s.Unlock()
	delete(s.connectedClients, clientID)
}
