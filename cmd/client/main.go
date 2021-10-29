package main

import (
	"flag"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
)

var (
	addr                  = flag.String("addr", "localhost:8080", "http service address")
	instrument            = flag.String("inst", "EURUSD", "instrument")
	interval              = flag.Duration("inter", 2*time.Second, "interval of sending request")
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	clientID              = seededRand.Uint32()
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go startRecieving(c, done)
	startFakingOrders(c, done, interrupt)
}

func startRecieving(c *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, mess, err := c.ReadMessage()
		if err != nil {
			log.Printf("recv error: %+v", err)
			return
		}
		log.Printf("recv: %v", proxy.DecodeOrderResponse(mess))
	}
}

func startFakingOrders(c *websocket.Conn, done chan struct{}, interrupt chan os.Signal) {
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	var id = uint32(time.Now().UTC().Unix())
	for {
		id++

		select {
		case <-done:
			return
		case <-ticker.C:
			req := proxy.OrderRequest{
				ClientID:   clientID,
				ID:         id,
				ReqType:    uint8(rand.Uint32()),
				OrderKind:  uint8(rand.Uint32()),
				Volume:     100 * seededRand.Float64(),
				Instrument: *instrument,
			}

			err := c.WriteMessage(websocket.TextMessage, proxy.EncodeOrderRequest(req))
			if err != nil {
				log.Printf("send err: %v", err)
				continue
			} else {
				log.Printf("sent: %v", req)
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			if err := c.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
