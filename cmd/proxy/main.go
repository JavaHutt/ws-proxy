package main

import (
	"flag"
	"log"

	"test.task/backend/proxy/internal/action"
	"test.task/backend/proxy/internal/adapter"
	"test.task/backend/proxy/internal/server"
	"test.task/backend/proxy/internal/service"
)

var (
	proxyAddr      = flag.String("proxyAddr", "localhost:8080", "http proxy address")
	backendAddr    = flag.String("addr", "localhost:8081", "http service address")
	ordersLimit    = flag.Uint("N", 4, "opened orders per client per instrument")
	volumeSumLimit = flag.Float64("S", 4400, "sum of volumes per client per instrument")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	orderAdapter := adapter.NewOrderAdapter()
	ordersService := service.NewOrdersService(*ordersLimit, *volumeSumLimit)
	server := server.NewServer(*proxyAddr, *backendAddr, orderAdapter, ordersService)
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})

	go func() {
		errorChannel <- server.Open()
	}()
	action.GracefulShutdown(errorChannel, server, doneChannel)
}
