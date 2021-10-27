package protocol

import (
	"encoding/binary"
	"math"
)

type (
	// OrderRequest ...
	OrderRequest struct {
		ClientID   uint32  // 4
		ID         uint32  // 4
		ReqType    uint8   // 1
		OrderKind  uint8   // 1
		Volume     float64 // 8
		Instrument string  // variadic
	}

	// OrderResponse ...
	OrderResponse struct {
		ID   uint32 // 4
		Code uint16 // 2
	}
)

var (
	bo        = binary.LittleEndian
	reqFixLen = 18
	resFixLen = 6
)

// EncodeOrderRequest ...
func EncodeOrderRequest(req OrderRequest) []byte {
	res := make([]byte, reqFixLen+len(req.Instrument))
	c := 0

	bo.PutUint32(res[c:c+4], req.ClientID)
	c += 4

	bo.PutUint32(res[c:c+4], req.ID)
	c += 4

	res[c] = req.ReqType
	c++

	res[c] = req.OrderKind
	c++

	PutFloat64(res[c:c+8], req.Volume)
	c += 8

	copy(res[c:], []byte(req.Instrument))
	return res
}

// DecodeOrderRequest decodes request
func DecodeOrderRequest(body []byte) OrderRequest {

	res := OrderRequest{}
	c := 0

	res.ClientID = bo.Uint32(body[c : c+4])
	c += 4

	res.ID = bo.Uint32(body[c : c+4])
	c += 4

	res.ReqType = body[c]
	c++

	res.OrderKind = body[c]
	c++

	res.Volume = Float64FromBytes(body[c : c+8])
	c += 8

	res.Instrument = string(body[c:])
	return res
}

// DecodeOrderResponse decodes request
func DecodeOrderResponse(body []byte) OrderResponse {
	res := OrderResponse{}
	c := 0

	res.ID = bo.Uint32(body[c : c+4])
	c += 4

	res.Code = bo.Uint16(body[c:])
	return res
}

func EncodeOrderResponse(resp OrderResponse) []byte {
	res := make([]byte, resFixLen, resFixLen)
	c := 0

	bo.PutUint32(res[c:c+4], resp.ID)
	c += 4

	bo.PutUint16(res[c:c+2], resp.Code)
	c += 2

	return res
}

func Float64FromBytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func PutFloat64(bytes []byte, float float64) {
	bits := math.Float64bits(float)
	binary.LittleEndian.PutUint64(bytes, bits)
}
