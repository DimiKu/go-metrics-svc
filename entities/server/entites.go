package server

const (
	CounterMetrics = "counter"
	GaugeMetrics   = "gauge"
)

// TODO кажется это не энтити , можно унести в репос
type StorageValue struct {
	Counter int64
	Gauge   float64
}
