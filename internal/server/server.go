package server

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
	"test.task/backend/proxy/internal/model"
)

// Server will perform operations over http.
type Server interface {
	// Open will setup a tcp listener and serve the http requests.
	Open() error

	// Close will close the socket if it's open.
	Close(ctx context.Context) error
}

type orderAdapter interface {
	TranslateOrder(order proxy.OrderRequest) (model.OrderRequest, error)
}

// Server represents an HTTP server.
type server struct {
	sync.Mutex
	addr             string
	backendAddr      string
	adapter          orderAdapter
	connectedClients map[uint32]struct{}
	serv             *http.Server
	upgrader         websocket.Upgrader
	dialer           *websocket.Dialer
}

func NewServer(addr, backendAddr string, orderAdapter orderAdapter) Server {
	return &server{
		addr:             addr,
		backendAddr:      backendAddr,
		adapter:          orderAdapter,
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
	_, err = s.adapter.TranslateOrder(req)
	if err != nil {
		// the task description didn't specify the way to respond to invalid
		// requests, so I've decided to send back "Other" result code
		writeOtherErrorToClient(clientWS, req.ID, err)
	} else {
		// first-time write to server after establishing connection with client
		writeToConn(serverWS, "server", mt, message)
	}

	done := make(chan struct{})
	// start listening from server and repeat message directly to client
	go s.serverToClient(serverWS, clientWS, done)
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

		_, err = s.adapter.TranslateOrder(req)
		if err != nil {
			writeOtherErrorToClient(clientWS, req.ID, err)
			continue
		}

		res := proxy.OrderResponse{
			ID:   req.ID,
			Code: 0,
		}
		if err = writeToConn(serverWS, "server", mt, message); err != nil {
			continue
		}

		log.Printf("sent to server: %v", res)
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

		if err = writeToConn(clientWS, "client", mt, messsage); err != nil {
			continue
		}

		log.Printf("recv from server and sent to client: %v", decoded)
	}
}

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
