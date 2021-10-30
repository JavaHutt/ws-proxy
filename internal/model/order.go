package model

type RequestType uint8

const (
	RequestTypeOpen RequestType = iota + 1
	RequestTypeClose
)

type OrderKind uint8

const (
	OrderKindBuy OrderKind = iota + 1
	OrderKindSell
)

type ResultCode uint16

const (
	ResultCodeSuccess ResultCode = iota
	ResultCodeOpenOrdersExceedes
	ResultCodeVolumesExceedes
	ResultCodeOther
)

// OrderRequest is the request from client to server
type OrderRequest struct {
	ClientID   uint32
	ID         uint32
	ReqType    RequestType
	OrderKind  OrderKind
	Volume     float64
	Instrument string
}

// OrderResponse is the response from server to client
type OrderResponse struct {
	ID   uint32
	Code ResultCode
}
