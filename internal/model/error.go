package model

import "errors"

type Error error

var (
	ErrInvalidRequest Error = errors.New("invalid request")
)
