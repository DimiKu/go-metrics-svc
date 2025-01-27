package custom_errors

import "errors"

var (
	ErrMetricNotExist = errors.New("metric not exist")
)
