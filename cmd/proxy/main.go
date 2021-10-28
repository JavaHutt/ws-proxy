package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
)

var (
	proxyAddr   = flag.String("proxyAddr", "localhost:8080", "http proxy address")
	backendAddr = flag.String("addr", "localhost:8081", "http service address")
	upgrader    = websocket.Upgrader{}
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", proxyHandler)
	log.Printf("Waiting for proxyions on %s/", *proxyAddr)
	log.Fatal(http.ListenAndServe(*proxyAddr, nil))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		req := proxy.DecodeOrderRequest(message)
		log.Printf("recv: %v", req)
		_ = req

		res := proxy.OrderResponse{
			ID:   req.ID,
			Code: 0,
		}
		err = c.WriteMessage(mt, proxy.EncodeOrderResponse(res))
		if err != nil {
			log.Println("write:", err)
			continue
		}

		log.Printf("sent: %v", req)
	}
}