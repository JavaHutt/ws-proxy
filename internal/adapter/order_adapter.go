package adapter

import (
	"fmt"

	proxy "test.task/backend/proxy"
	"test.task/backend/proxy/internal/model"
)

type orderAdapter struct{}

func NewOrderAdapter() *orderAdapter {
	return &orderAdapter{}
}

func (orderAdapter) TranslateOrder(order proxy.OrderRequest) (model.OrderRequest, error) {
	if err := validate(order); err != nil {
		return model.OrderRequest{}, err
	}
	return model.OrderRequest{
		ClientID:   order.ClientID,
		ID:         order.ID,
		ReqType:    model.RequestType(order.ReqType),
		OrderKind:  model.OrderKind(order.OrderKind),
		Volume:     order.Volume,
		Instrument: order.Instrument,
	}, nil
}

func (orderAdapter) GetResultCodeFromErr(err error) model.ResultCode {
	switch err {
	case model.ErrNumberExceedes:
		return model.ResultCodeOpenOrdersExceedes
	case model.ErrVolumeSumExceedes:
		return model.ResultCodeVolumesExceedes
	default:
		return model.ResultCodeOther
	}
}

func validate(order proxy.OrderRequest) error {
	reqType, orderKind := order.ReqType, order.OrderKind

	if reqType < uint8(model.RequestTypeOpen) || reqType > uint8(model.RequestTypeClose) {
		return fmt.Errorf("%w: invalid request type", model.ErrInvalidRequest)
	}
	if orderKind < uint8(model.OrderKindBuy) || orderKind > uint8(model.OrderKindSell) {
		return fmt.Errorf("%w: invalid order kind", model.ErrInvalidRequest)
	}

	return nil
}
