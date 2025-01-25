package server

const (
	CounterMetrics = "counter"
	GaugeMetrics   = "gauge"
)

type StorageValue struct {
	Counter int64
	Gauge   float64
}
