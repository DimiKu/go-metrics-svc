package dto

// Dto для хендлеров

var (
	MetricTypeHandlerCounterTypeDto = MetricTypeServiceCounterTypeDto
	MetricTypeHandlerGaugeTypeDto   = MetricTypeServiceGaugeTypeDto
)

type MetricHandlerDto MetricServiceDto
