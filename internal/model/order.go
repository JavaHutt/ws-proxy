package model

type RequestType int

const (
	RequestTypeOpen RequestType = iota + 1
	RequestTypeClose
)

type OrderKind int

const (
	OrderKindBuy OrderKind = iota + 1
	OrderKindSell
)

type ResultCode int

const (
	ResultCodeOpenOrdersExceedes ResultCode = iota + 1
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
