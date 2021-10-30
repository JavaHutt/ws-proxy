package action

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test.task/backend/proxy/internal/server"
)

func GracefulShutdown(
	errorChannel chan error,
	httpServer server.Server,
	doneChannel chan struct{},
) {
	// Capture interrupts.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errorChannel <- fmt.Errorf("got signal: %s %v", <-c, time.Now())
	}()

	if err := <-errorChannel; err != nil {
		log.Println(err)
		close(doneChannel)
		httpServerShutdown(httpServer)

		log.Println("app stopped", time.Now())
	}
	os.Exit(1)
}

func httpServerShutdown(httpServer server.Server) {
	ctx := context.Background()

	if err := httpServer.Close(ctx); err != nil {
		log.Printf("could not gracefully shutdown the server: %v\n", err)
	}

	log.Println("http server stopped", time.Now())
}
