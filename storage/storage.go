package storage

import (
	"go.uber.org/zap"
)

type StorageValue struct {
	Counter int64
	Gauge   float64
}

type MemStorage struct {
	metricsMap map[string]StorageValue

	log *zap.Logger
}

func NewMemStorage(metricsMap map[string]StorageValue, log *zap.Logger) *MemStorage {
	return &MemStorage{
		metricsMap: metricsMap,
		log:        log,
	}
}

func (m *MemStorage) UpdateValue(metricName string, metricValue float64) {
	m.log.Info("Update in storage")
	m.metricsMap[metricName] = StorageValue{Gauge: metricValue}
}

func (m *MemStorage) SumValue(metricName string, metricValue int64) {
	if value, exists := m.metricsMap[metricName]; exists {
		m.metricsMap[metricName] = StorageValue{
			Counter: value.Counter + metricValue,
		}
	} else {
		m.metricsMap[metricName] = StorageValue{
			Counter: metricValue,
		}
	}
}
