package protocol

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestOrderRequestEncode(t *testing.T) {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			want := OrderRequest{
				ClientID:   seededRand.Uint32(),
				ID:         seededRand.Uint32(),
				ReqType:    uint8(rand.Uint32()),
				OrderKind:  uint8(rand.Uint32()),
				Volume:     rand.Float64(),
				Instrument: RandomString(5, 7),
			}

			bytes := EncodeOrderRequest(want)
			got := DecodeOrderRequest(bytes)

			if got != want {
				t.Fatalf("OrderRequest mismatch: %+v", want)
			}
		})
	}
}

func TestOrderResponseEncode(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			want := OrderResponse{
				ID:   seededRand.Uint32(),
				Code: uint16(seededRand.Uint32()),
			}

			bytes := EncodeOrderResponse(want)
			got := DecodeOrderResponse(bytes)

			if got != want {
				t.Fatalf("OrderRequest mismatch: %+v", want)
			}
		})
	}
}

func StringWithCharset(length int, charset string) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(min, max int) string {
	if max < min {
		max, min = min, max
	}

	len := min + int(seededRand.Int31n(int32(max-min)))
	return String(len)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
