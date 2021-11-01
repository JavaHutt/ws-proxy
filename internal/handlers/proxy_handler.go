package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
	"test.task/backend/proxy/internal/model"
)

type orderAdapter interface {
	TranslateOrder(order proxy.OrderRequest) (model.OrderRequest, error)
	GetResultCodeFromErr(err error) model.ResultCode
}

type ordersService interface {
	ProcessOrder(order model.OrderRequest) error
}

type ProxyHandler struct {
	sync.Mutex
	backendAddr      string
	adapter          orderAdapter
	svc              ordersService
	connectedClients map[uint32]struct{}
	upgrader         websocket.Upgrader
	dialer           *websocket.Dialer
}

func NewProxyHandler(backendAddr string, adapter orderAdapter, svc ordersService) *ProxyHandler {
	return &ProxyHandler{
		backendAddr:      backendAddr,
		adapter:          adapter,
		svc:              svc,
		connectedClients: make(map[uint32]struct{}),
		upgrader:         websocket.Upgrader{},
		dialer:           websocket.DefaultDialer,
	}
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientWS, err := p.upgrader.Upgrade(w, r, nil)
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
	filterPassed := p.filterConnection(clientWS, clientID)
	if !filterPassed {
		return
	}

	serverWS := p.mustGetServerConn()
	// firstProcess runs once when connection had been established
	p.firstProcess(clientWS, serverWS, req, mt, message)

	// start listening from server and repeat message directly to client
	go p.serverToClient(serverWS, clientWS)
	// process client message and pass it to server if everything is ok
	p.clientToServer(clientWS, serverWS, clientID)
}

func (p *ProxyHandler) clientToServer(clientWS, serverWS *websocket.Conn, clientID uint32) {
	defer clientWS.Close()
	for {
		mt, message, err := clientWS.ReadMessage()
		if err != nil {
			p.disconnectClient(clientID)
			break
		}
		req := proxy.DecodeOrderRequest(message)
		log.Printf("recv from client: %v", req)
		id := req.ID

		translatedOrder, err := p.adapter.TranslateOrder(req)
		if err != nil {
			p.writeErrorToClient(clientWS, id, err)
			continue
		}
		if err := p.svc.ProcessOrder(translatedOrder); err != nil {
			p.writeErrorToClient(clientWS, id, err)
			continue
		}

		if err = writeToConn(serverWS, "server", mt, message); err != nil {
			continue
		}

		log.Printf("sent to server: %v", req)
	}
}

func (p *ProxyHandler) serverToClient(serverWS, clientWS *websocket.Conn) {
	defer clientWS.Close()
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

func (p *ProxyHandler) firstProcess(
	clientWS, serverWS *websocket.Conn,
	req proxy.OrderRequest,
	mt int,
	message []byte,
) {
	id := req.ID
	translatedOrder, err := p.adapter.TranslateOrder(req)
	if err != nil {
		// the task description didn't specify the way to respond to invalid
		// requests, so I've decided to send back "Other" result code
		p.writeErrorToClient(clientWS, id, err)
		return
	}
	if err = p.svc.ProcessOrder(translatedOrder); err != nil {
		p.writeErrorToClient(clientWS, id, err)
		return
	}
	// first-time write to server after establishing connection with client
	writeToConn(serverWS, "server", mt, message)
}
