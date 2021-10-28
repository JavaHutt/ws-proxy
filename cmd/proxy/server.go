package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
)

// WebSocketConnection is a wrapper for our websocket connection, in case
// we ever need to put more data into the struct
type WebSocketConnection struct {
	*websocket.Conn
}

// WsPayload defines the websocket request from the client
type WsPayload struct {
	Conn WebSocketConnection `json:"-"`
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	clientWS, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade client request:", err)
		return
	}

	u := url.URL{Scheme: "ws", Host: *backendAddr, Path: "/connect"}
	serverWS, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial to a server:", err)
	}

	done := make(chan struct{})
	go startRecievingFromServerToClient(serverWS, clientWS, done)

	go func() {
		defer clientWS.Close()
		for {
			mt, message, err := clientWS.ReadMessage()
			if err != nil {
				break
			}
			req := proxy.DecodeOrderRequest(message)
			log.Printf("recv from client: %v", req)

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
	}()
}

func startRecievingFromServerToClient(serverWS, clientWS *websocket.Conn, done chan struct{}) {
	for {
		defer serverWS.Close()
		mt, messsage, err := serverWS.ReadMessage()
		if err != nil {
			log.Printf("recv error: %+v", err)
			return
		}
		decoded := proxy.DecodeOrderResponse(messsage)
		log.Printf("recv from server: %v", decoded)

		if err = clientWS.WriteMessage(mt, messsage); err != nil {
			log.Println("write to client:", err)
			continue
		}

		log.Printf("sent to client: %v", decoded)
	}
}
