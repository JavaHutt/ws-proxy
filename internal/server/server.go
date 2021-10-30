package server

import (
	"context"
	"log"
	"net/http"
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
	GetResultCodeFromErr(err error) model.ResultCode
}

type ordersService interface {
	ProcessOrder(order model.OrderRequest) error
}

// Server represents an HTTP server.
type server struct {
	sync.Mutex
	addr             string
	backendAddr      string
	adapter          orderAdapter
	svc              ordersService
	connectedClients map[uint32]struct{}
	serv             *http.Server
	upgrader         websocket.Upgrader
	dialer           *websocket.Dialer
}

func NewServer(addr, backendAddr string, orderAdapter orderAdapter, ordersService ordersService) Server {
	return &server{
		addr:             addr,
		backendAddr:      backendAddr,
		adapter:          orderAdapter,
		svc:              ordersService,
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
