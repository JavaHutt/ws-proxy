package model

import "errors"

type Error error

var (
	ErrInvalidRequest    Error = errors.New("invalid request")
	ErrNumberExceedes    Error = errors.New("number of open orders exceeds")
	ErrVolumeSumExceedes Error = errors.New("sum volumes of orders exceeds")
	ErrNoOrderToClose    Error = errors.New("no order to close")
	ErrNegativeVolumeSum Error = errors.New("negative volume sum violation")
)
