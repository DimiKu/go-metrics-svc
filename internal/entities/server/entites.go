package server

const (
	CounterMetrics = "counter"
	GaugeMetrics   = "gauge"
)

type Metric struct {
	Name       string
	MetricType string
	Value      string
}
