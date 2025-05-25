package dto

import (
	"go-metric-svc/internal/entities/server"
)

// Dto для слоя сервиса

var (
	MetricTypeServiceCounterTypeDto = server.CounterMetrics
	MetricTypeServiceGaugeTypeDto   = server.GaugeMetrics
)

type MetricServiceDto server.Metric

type MetricCollectionDto struct {
	GaugeCollection   []MetricServiceDto
	CounterCollection []MetricServiceDto
}
