package adapter

import (
	"errors"
	"testing"

	proxy "test.task/backend/proxy"
	"test.task/backend/proxy/internal/model"
)

var mockAdapter = orderAdapter{}

func TestTranslateOrder(t *testing.T) {
	cases := []struct {
		name    string
		input   proxy.OrderRequest
		want    model.OrderRequest
		wantErr error
	}{
		{
			name: "valid request",
			input: proxy.OrderRequest{
				ClientID:   1,
				ID:         1,
				ReqType:    1,
				OrderKind:  1,
				Volume:     1000,
				Instrument: "USDRUB",
			},
			want: model.OrderRequest{
				ClientID:   1,
				ID:         1,
				ReqType:    model.RequestTypeOpen,
				OrderKind:  model.OrderKindBuy,
				Volume:     1000,
				Instrument: "USDRUB",
			},
			wantErr: nil,
		},
		{
			name: "invalid request type",
			input: proxy.OrderRequest{
				ClientID:   1,
				ID:         1,
				ReqType:    123,
				OrderKind:  1,
				Volume:     1000,
				Instrument: "USDRUB",
			},
			want:    model.OrderRequest{},
			wantErr: model.ErrInvalidRequest,
		},
		{
			name: "invalid request kind",
			input: proxy.OrderRequest{
				ClientID:   1,
				ID:         1,
				ReqType:    1,
				OrderKind:  123,
				Volume:     1000,
				Instrument: "USDRUB",
			},
			want:    model.OrderRequest{},
			wantErr: model.ErrInvalidRequest,
		},
	}
	for _, tc := range cases {
		got, err := mockAdapter.TranslateOrder(tc.input)
		if got != tc.want {
			t.Fatalf("%s failed: expected: %v, got: %v", tc.name, tc.want, got)
		}
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s failed: expected err: %v, got: %v", tc.name, tc.wantErr, err)
		}
	}
}
