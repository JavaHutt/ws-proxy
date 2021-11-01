package main

import (
	"flag"
	"log"

	"test.task/backend/proxy/internal/action"
	"test.task/backend/proxy/internal/adapter"
	"test.task/backend/proxy/internal/handlers"
	"test.task/backend/proxy/internal/http"
	"test.task/backend/proxy/internal/service"
)

var (
	addr           = flag.String("addr", "localhost:8080", "http proxy address")
	backendAddr    = flag.String("backendAddr", "localhost:8081", "http service address")
	ordersLimit    = flag.Uint("N", 4, "opened orders per client per instrument")
	volumeSumLimit = flag.Float64("S", 4400, "sum of volumes per client per instrument")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	orderAdapter := adapter.NewOrderAdapter()
	ordersService := service.NewOrdersService(*ordersLimit, *volumeSumLimit)
	proxyHandler := handlers.NewProxyHandler(*backendAddr, orderAdapter, ordersService)
	server := http.NewServer(*addr, proxyHandler)

	errorChannel := make(chan error)
	doneChannel := make(chan struct{})

	go func() {
		errorChannel <- server.Open()
	}()
	action.GracefulShutdown(errorChannel, server, doneChannel)
}
