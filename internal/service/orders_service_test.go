package service

import (
	"errors"
	"testing"

	"test.task/backend/proxy/internal/model"
)

func TestProcessOrder(t *testing.T) {
	reqTypeOpen := model.RequestTypeOpen
	reqTypeClose := model.RequestTypeClose
	clientID := uint32(1)
	instrumentName := "USDRUB"

	cases := []struct {
		name       string
		service    *ordersService
		input      model.OrderRequest
		wantCount  uint
		wantVolume float64
		wantErr    error
	}{
		{
			name:    "invalid request type",
			service: NewOrdersService(1, 200),
			input: model.OrderRequest{
				ReqType: 4,
			},
			wantErr: model.ErrInvalidRequest,
		},
		{
			name:    "open order success",
			service: NewOrdersService(1, 200),
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Volume:     100,
				Instrument: instrumentName,
			},
			wantCount:  1,
			wantVolume: 100,
			wantErr:    nil,
		},
		{
			name: "increase volume and count",
			service: &ordersService{
				ordersLimit:    2,
				volumeSumLimit: 4000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {
						instrumentName: {
							count:     1,
							volumeSum: 2500,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Volume:     1000,
				Instrument: instrumentName,
			},
			wantCount:  2,
			wantVolume: 3500,
			wantErr:    nil,
		},
		{
			name: "instrument not exists on client",
			service: &ordersService{
				ordersLimit:    2,
				volumeSumLimit: 4000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Volume:     1000,
				Instrument: instrumentName,
			},
			wantCount:  1,
			wantVolume: 1000,
			wantErr:    nil,
		},
		{
			name:    "open order with restricted limit",
			service: NewOrdersService(0, 100),
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Instrument: instrumentName,
			},
			wantErr: model.ErrNumberExceedes,
		},
		{
			name: "open order with number of orders exceedes",
			service: &ordersService{
				ordersLimit:    2,
				volumeSumLimit: 1000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {
						instrumentName: {
							count: 2,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Instrument: instrumentName,
			},
			wantCount: 3,
			wantErr:   model.ErrNumberExceedes,
		},
		{
			name:    "open order with restricted sum limit",
			service: NewOrdersService(10, 0),
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Volume:     100,
				Instrument: instrumentName,
			},
			wantErr: model.ErrVolumeSumExceedes,
		},
		{
			name: "open order with sum of volumes exceedes",
			service: &ordersService{
				ordersLimit:    2,
				volumeSumLimit: 3000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {
						instrumentName: {
							count:     1,
							volumeSum: 2500,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Volume:     1000,
				Instrument: instrumentName,
			},
			wantErr: model.ErrVolumeSumExceedes,
		},
		{
			name: "close order success",
			service: &ordersService{
				ordersLimit:    5,
				volumeSumLimit: 4000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {
						instrumentName: {
							count:     3,
							volumeSum: 2500,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeClose,
				Volume:     100,
				Instrument: instrumentName,
			},
			wantCount:  2,
			wantVolume: 2400,
			wantErr:    nil,
		},
		{
			name: "close order no client",
			service: &ordersService{
				clientsInstruments: map[uint32]map[string]*instrument{},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeClose,
				Volume:     1000,
				Instrument: instrumentName,
			},
			wantErr: model.ErrNoOrderToClose,
		},
		{
			name: "close order no instrument",
			service: &ordersService{
				ordersLimit:    2,
				volumeSumLimit: 4000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeClose,
				Volume:     1000,
				Instrument: instrumentName,
			},
			wantErr: model.ErrNoOrderToClose,
		},
		{
			name: "close order zero orders",
			service: &ordersService{
				ordersLimit:    2,
				volumeSumLimit: 4000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {
						instrumentName: {
							count: 0,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeClose,
				Instrument: instrumentName,
			},
			wantErr: model.ErrNoOrderToClose,
		},
		{
			name: "close order negative volume sum",
			service: &ordersService{
				ordersLimit:    5,
				volumeSumLimit: 4000,
				clientsInstruments: map[uint32]map[string]*instrument{
					clientID: {
						instrumentName: {
							count:     3,
							volumeSum: 300,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeClose,
				Instrument: instrumentName,
				Volume:     400,
			},
			wantErr: model.ErrNegativeVolumeSum,
		},
	}
	for _, tc := range cases {
		err := tc.service.ProcessOrder(tc.input)
		if err == nil {
			instr := tc.service.clientsInstruments[clientID][instrumentName]
			if instr.count != tc.wantCount {
				t.Fatalf("%s failed: expected count: %d, got: %d",
					tc.name, tc.wantCount, instr.count)
			}
			if instr.volumeSum != tc.wantVolume {
				t.Fatalf("%s failed: expected volume: %f, got: %f",
					tc.name, tc.wantVolume, instr.volumeSum)
			}
		}

		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s failed: expected err: %v, got: %v", tc.name, tc.wantErr, err)
		}
	}
}
