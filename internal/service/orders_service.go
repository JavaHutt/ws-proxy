package service

import (
	"sync"

	"test.task/backend/proxy/internal/model"
)

type instrument struct {
	count     uint
	volumeSum float64
}

type orderService struct {
	// I've decided to use map + mutex instead of syncmap because there are
	// gonna be constant key manipulations operations since it is client id
	sync.Mutex
	ordersLimit        uint
	volumeSumLimit     float64
	clientsInstruments map[uint32]map[string]instrument
}

func NewOrderService(ordersLimit uint, volumeSumLimit float64) *orderService {
	return &orderService{
		ordersLimit:        ordersLimit,
		volumeSumLimit:     volumeSumLimit,
		clientsInstruments: make(map[uint32]map[string]instrument),
	}
}

func (svc *orderService) OpenOrder(order model.OrderRequest) error {
	if svc.ordersLimit <= 0 {
		return model.ErrNumberExceedes
	}
	clientID, orderInstrument, volume := order.ClientID, order.Instrument, order.Volume
	if volume > svc.volumeSumLimit {
		return model.ErrVolumeSumExceedes
	}

	svc.Lock()
	defer svc.Unlock()
	instrumentMap, clientExists := svc.clientsInstruments[clientID]
	if !clientExists {
		svc.clientsInstruments[clientID] = make(map[string]instrument)
		svc.clientsInstruments[clientID][orderInstrument] = instrument{
			count:     1,
			volumeSum: volume,
		}
		return nil
	}
	instr, instrumentExist := instrumentMap[orderInstrument]
	if !instrumentExist {
		instrumentMap[orderInstrument] = instrument{
			count:     1,
			volumeSum: volume,
		}
		return nil
	}

	if instr.count+1 > svc.ordersLimit {
		return model.ErrNumberExceedes
	}
	if instr.volumeSum+volume > svc.volumeSumLimit {
		return model.ErrVolumeSumExceedes
	}
	instr.count++
	instr.volumeSum += volume

	return nil
}

func (orderService) CloseOrder() {

}
