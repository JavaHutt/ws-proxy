package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	proxyAddr   = flag.String("proxyAddr", "localhost:8080", "http proxy address")
	backendAddr = flag.String("addr", "localhost:8081", "http service address")
	upgrader    = websocket.Upgrader{}
	dialer      = websocket.DefaultDialer
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", proxyHandler)
	log.Printf("Waiting for proxyions on %s/", *proxyAddr)
	log.Fatal(http.ListenAndServe(*proxyAddr, nil))
}
