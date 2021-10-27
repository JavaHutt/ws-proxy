package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/koding/websocketproxy"
)

var (
	proxyAddr = flag.String("proxyAddr", "localhost:8080", "http proxy address")
	addr      = flag.String("addr", "localhost:8081", "http service address")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	u := url.URL{Scheme: "ws", Host: *addr}

	proxy := websocketproxy.NewProxy(&u)
	http.HandleFunc("/", proxyHandler(proxy))
	log.Fatal(http.ListenAndServe(*proxyAddr, nil))
}

func proxyHandler(proxy *websocketproxy.WebsocketProxy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("proxy req")
		proxy.ServeHTTP(w, r)
	}
}
