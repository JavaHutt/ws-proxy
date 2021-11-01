package service

import (
	"sync"
)

type clientsService struct {
	sync.Mutex
	connectedClients map[uint32]struct{}
}

func NewClientsService() *clientsService {
	return &clientsService{
		connectedClients: make(map[uint32]struct{}),
	}
}

// TryConnectClient tries to connect a new client
// and returns true if operation was successful
func (svc *clientsService) TryConnectClient(clientID uint32) bool {
	svc.Lock()
	defer svc.Unlock()
	if _, ok := svc.connectedClients[clientID]; ok {
		return false
	}
	// connecting new client
	svc.connectedClients[clientID] = struct{}{}
	return true
}

func (svc *clientsService) DisconnectClient(clientID uint32) {
	svc.Lock()
	defer svc.Unlock()
	delete(svc.connectedClients, clientID)
}
