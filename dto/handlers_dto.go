package dto

// Dto для хендлеров. Используется для передачи данных между слоями хендлера и сервиса
var (
	MetricTypeHandlerCounterTypeDto = MetricTypeServiceCounterTypeDto
	MetricTypeHandlerGaugeTypeDto   = MetricTypeServiceGaugeTypeDto
)

type MetricHandlerDto MetricServiceDto
