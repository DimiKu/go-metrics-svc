package dto

import (
	"go-metric-svc/internal/entities/server"
)

var (
	MetricTypeServiceCounterTypeDto = server.CounterMetrics
	MetricTypeServiceGaugeTypeDto   = server.GaugeMetrics
)

type MetricServiceDto server.Metric

type MetricCollectionDto struct {
	GaugeCollection   []MetricServiceDto
	CounterCollection []MetricServiceDto
}
