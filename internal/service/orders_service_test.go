package service

import (
	"errors"
	"testing"

	"test.task/backend/proxy/internal/model"
)

func TestOpenOrder(t *testing.T) {
	reqTypeOpen := model.RequestTypeOpen
	clientID := uint32(1)
	instrumentName := "USDRUB"

	cases := []struct {
		name       string
		service    *orderService
		input      model.OrderRequest
		wantCount  uint
		wantVolume float64
		wantErr    error
	}{
		{
			name:    "open order success",
			service: NewOrderService(1, 200),
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
			service: &orderService{
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
			service: &orderService{
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
			service: NewOrderService(0, 100),
			input: model.OrderRequest{
				ClientID:   clientID,
				ReqType:    reqTypeOpen,
				Instrument: instrumentName,
			},
			wantErr: model.ErrNumberExceedes,
		},
		{
			name: "open order with number of orders exceedes",
			service: &orderService{
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
			service: NewOrderService(10, 0),
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
			service: &orderService{
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
	}
	for _, tc := range cases {
		err := tc.service.OpenOrder(tc.input)
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
