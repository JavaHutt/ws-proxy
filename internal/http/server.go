package http

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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
	handler          http.Handler
	addr             string
	connectedClients map[uint32]struct{}
	serv             *http.Server
	upgrader         websocket.Upgrader
	dialer           *websocket.Dialer
}

func NewServer(addr string, handler http.Handler) Server {
	return &server{
		handler:          handler,
		addr:             addr,
		connectedClients: make(map[uint32]struct{}),
		upgrader:         websocket.Upgrader{},
		dialer:           websocket.DefaultDialer,
	}
}

// Open will setup a tcp listener and serve the http requests.
func (s *server) Open() error {
	s.serv = &http.Server{
		Addr:    s.addr,
		Handler: s.handler,
	}
	log.Printf("Waiting for connections on %s/", s.addr)
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
