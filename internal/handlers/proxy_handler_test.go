package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	proxy "test.task/backend/proxy"
	"test.task/backend/proxy/internal/adapter"
	"test.task/backend/proxy/internal/service"
)

func TestProxyHandler(t *testing.T) {
	cases := []struct {
		name          string
		ordersService ordersService
		request       proxy.OrderRequest
		response      proxy.OrderResponse
	}{
		{
			name:          "open order successful",
			ordersService: service.NewOrdersService(4, 3000),
			request: proxy.OrderRequest{
				ClientID:   4815,
				ID:         162342,
				ReqType:    1,
				OrderKind:  1,
				Volume:     1000,
				Instrument: "USDEUR",
			},
			response: proxy.OrderResponse{
				ID:   162342,
				Code: 0,
			},
		},
		{
			name:          "open orders exceedes",
			ordersService: service.NewOrdersService(0, 500),
			request: proxy.OrderRequest{
				ClientID:   4815,
				ID:         162342,
				ReqType:    1,
				OrderKind:  1,
				Volume:     50,
				Instrument: "USDEUR",
			},
			response: proxy.OrderResponse{
				ID:   162342,
				Code: 1,
			},
		},
		{
			name:          "sum of volumes exceedes",
			ordersService: service.NewOrdersService(4, 500),
			request: proxy.OrderRequest{
				ClientID:   4815,
				ID:         162342,
				ReqType:    1,
				OrderKind:  1,
				Volume:     5000,
				Instrument: "USDEUR",
			},
			response: proxy.OrderResponse{
				ID:   162342,
				Code: 2,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewProxyHandler(
				"localhost:8081",
				adapter.NewOrderAdapter(),
				tc.ordersService,
				service.NewClientsService(),
			)

			s, ws := newWSServer(t, handler)
			defer s.Close()
			defer ws.Close()

			sendMessage(t, ws, tc.request)

			got := receiveWSMessage(t, ws)

			if got != tc.response {
				t.Fatalf("Expected %+v, got %+v", tc.response, got)
			}
		})
	}
}

func newWSServer(t *testing.T, h http.Handler) (*httptest.Server, *websocket.Conn) {
	t.Helper()

	s := httptest.NewServer(h)
	wsURL := httpToWS(t, s.URL)

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	return s, ws
}

func sendMessage(t *testing.T, ws *websocket.Conn, msg proxy.OrderRequest) {
	t.Helper()

	if err := ws.WriteMessage(websocket.BinaryMessage, proxy.EncodeOrderRequest(msg)); err != nil {
		t.Fatalf("%v", err)
	}
}

func receiveWSMessage(t *testing.T, ws *websocket.Conn) proxy.OrderResponse {
	t.Helper()

	_, m, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	res := proxy.DecodeOrderResponse(m)

	return res
}

func httpToWS(t *testing.T, u string) string {
	t.Helper()

	wsURL, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}

	switch wsURL.Scheme {
	case "http":
		wsURL.Scheme = "ws"
	case "https":
		wsURL.Scheme = "wss"
	}

	return wsURL.String()
}
