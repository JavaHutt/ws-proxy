package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	proxyAddr   = flag.String("proxyAddr", "localhost:8080", "http proxy address")
	backendAddr = flag.String("addr", "localhost:8081", "http service address")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", handlerProxy)
	log.Printf("Proxy is waiting for connections on %s/", *proxyAddr)
	log.Fatal(http.ListenAndServe(*proxyAddr, nil))
}

func handlerProxy(w http.ResponseWriter, r *http.Request) {
	url := url.URL{Scheme: "http", Host: *backendAddr}
	proxy := httputil.NewSingleHostReverseProxy(&url)
	proxy.ServeHTTP(w, r)
}
