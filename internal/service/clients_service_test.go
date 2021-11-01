package service

import (
	"testing"
)

func TestTryConnectClient(t *testing.T) {
	clientID := 2342

	cases := []struct {
		name    string
		client  uint32
		service *clientsService
		want    bool
	}{
		{
			name:    "connect client success",
			client:  uint32(clientID),
			service: NewClientsService(),
			want:    true,
		},
		{
			name:   "client already connected",
			client: uint32(clientID),
			service: &clientsService{
				connectedClients: map[uint32]struct{}{
					481516: struct{}{},
					2342:   struct{}{},
				},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		connected := tc.service.TryConnectClient(tc.client)
		if connected != tc.want {
			t.Fatalf("%s failed: expected result: %t, got: %t",
				tc.name, tc.want, connected)
		}
	}
}
