package dto

import (
	"go-metric-svc/entities/server"
)

var (
	MetricTypeServiceCounterTypeDto = server.CounterMetrics
	MetricTypeServiceGaugeTypeDto   = server.GaugeMetrics
)
