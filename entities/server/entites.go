package server

const (
	Int64Metrics   = "int64"
	CounterMetrics = "counter"
	GaugeMetrics   = "gauge"
	Float64Metrics = "float64"
)

var SumMetrics = []string{Int64Metrics, CounterMetrics}
var UpdateMetrics = []string{GaugeMetrics, Float64Metrics}
