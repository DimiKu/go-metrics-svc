package customErrors

import "errors"

var (
	ErrMetricNotExist = errors.New("metric not exist")
)
