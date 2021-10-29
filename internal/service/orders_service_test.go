package service

import (
	"errors"
	"testing"

	"test.task/backend/proxy/internal/model"
)

func TestOpenOrder(t *testing.T) {
	reqTypeOpen := model.RequestTypeOpen
	cases := []struct {
		name    string
		service *orderService
		input   model.OrderRequest
		wantErr error
	}{
		{
			name:    "open order success",
			service: NewOrderService(1, 200),
			input: model.OrderRequest{
				ClientID:   1,
				ReqType:    reqTypeOpen,
				Volume:     100,
				Instrument: "USDRUB",
			},
			wantErr: nil,
		},
		{
			name:    "open order with restricted limit",
			service: NewOrderService(0, 100),
			input: model.OrderRequest{
				ClientID:   1,
				ReqType:    reqTypeOpen,
				Instrument: "USDRUB",
			},
			wantErr: model.ErrNumberExceedes,
		},
		{
			name: "open order with number of orders exceedes",
			service: &orderService{
				ordersLimit:    2,
				volumeSumLimit: 1000,
				clientsInstruments: map[uint32]map[string]instrument{
					1: {
						"USDRUB": {
							count: 2,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   1,
				ReqType:    reqTypeOpen,
				Instrument: "USDRUB",
			},
			wantErr: model.ErrNumberExceedes,
		},
		{
			name:    "open order with restricted sum limit",
			service: NewOrderService(10, 0),
			input: model.OrderRequest{
				ClientID:   1,
				ReqType:    reqTypeOpen,
				Volume:     100,
				Instrument: "USDRUB",
			},
			wantErr: model.ErrVolumeSumExceedes,
		},
		{
			name: "open order with sum of volumes exceedes",
			service: &orderService{
				ordersLimit:    2,
				volumeSumLimit: 3000,
				clientsInstruments: map[uint32]map[string]instrument{
					1: {
						"XLMEUR": {
							count:     1,
							volumeSum: 2500,
						},
					},
				},
			},
			input: model.OrderRequest{
				ClientID:   1,
				ReqType:    reqTypeOpen,
				Volume:     1000,
				Instrument: "XLMEUR",
			},
			wantErr: model.ErrVolumeSumExceedes,
		},
	}
	for _, tc := range cases {
		err := tc.service.OpenOrder(tc.input)
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s failed: expected err: %v, got: %v", tc.name, tc.wantErr, err)
		}
	}
}
