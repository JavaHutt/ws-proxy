package service

import (
	"sync"

	"log"

	"test.task/backend/proxy/internal/model"
)

type instrument struct {
	count     uint
	volumeSum float64
}

type ordersService struct {
	// I've decided to use map + mutex instead of syncmap because there are
	// gonna be constant key manipulations we can have many clients
	sync.Mutex
	ordersLimit        uint
	volumeSumLimit     float64
	clientsInstruments map[uint32]map[string]*instrument
}

func NewOrdersService(ordersLimit uint, volumeSumLimit float64) *ordersService {
	log.Printf("orders service started. open orders limit: %d, sum of volumes limit: %f\n", ordersLimit, volumeSumLimit)

	return &ordersService{
		ordersLimit:        ordersLimit,
		volumeSumLimit:     volumeSumLimit,
		clientsInstruments: make(map[uint32]map[string]*instrument),
	}
}

// ProcessOrder is an entry point in orders service
func (svc *ordersService) ProcessOrder(order model.OrderRequest) error {
	switch order.ReqType {
	case model.RequestTypeOpen:
		return svc.openOrder(order)
	case model.RequestTypeClose:
		return svc.closeOrder(order)
	default:
		return model.ErrInvalidRequest
	}
}

func (svc *ordersService) openOrder(order model.OrderRequest) error {
	if svc.ordersLimit == 0 {
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
		svc.clientsInstruments[clientID] = make(map[string]*instrument)
		svc.clientsInstruments[clientID][orderInstrument] = &instrument{
			count:     1,
			volumeSum: volume,
		}
		return nil
	}
	instr, instrumentExist := instrumentMap[orderInstrument]
	if !instrumentExist {
		instrumentMap[orderInstrument] = &instrument{
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

func (svc *ordersService) closeOrder(order model.OrderRequest) error {
	clientID, orderInstrument, volume := order.ClientID, order.Instrument, order.Volume

	svc.Lock()
	defer svc.Unlock()
	instrumentMap, clientExists := svc.clientsInstruments[clientID]
	if !clientExists {
		return model.ErrNoOrderToClose
	}
	instr, instrumentExist := instrumentMap[orderInstrument]
	if !instrumentExist {
		return model.ErrNoOrderToClose
	}
	if instr.count == 0 {
		return model.ErrNoOrderToClose
	}
	if instr.volumeSum-volume < 0 {
		return model.ErrNegativeVolumeSum
	}
	instr.count--
	instr.volumeSum -= volume

	return nil
}
