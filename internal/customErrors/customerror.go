package customerrors

import "errors"

var (
	ErrMetricNotExist = errors.New("metric not exist")
	ErrHashMissMatch  = errors.New("hash miss match")
)
