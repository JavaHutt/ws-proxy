package http

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
