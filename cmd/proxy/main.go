package main

import (
	"flag"
	"log"

	"test.task/backend/proxy/internal/action"
	"test.task/backend/proxy/internal/adapter"
	"test.task/backend/proxy/internal/server"
)

var (
	proxyAddr   = flag.String("proxyAddr", "localhost:8080", "http proxy address")
	backendAddr = flag.String("addr", "localhost:8081", "http service address")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	orderAdapter := adapter.NewOrderAdapter()
	server := server.NewServer(*proxyAddr, *backendAddr, orderAdapter)
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})

	go func() {
		errorChannel <- server.Open()
	}()
	action.GracefulShutdown(errorChannel, server, doneChannel)
}
