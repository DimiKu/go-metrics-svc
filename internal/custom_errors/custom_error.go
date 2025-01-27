package custom_errors

import "errors"

var (
	MetricNotExist = errors.New("metric not exist")
)
