package storage

import (
	"fmt"
	"go-metric-svc/entities/server"
	"go.uber.org/zap"
)

type MemStorage struct {
	metricsMap map[string]server.StorageValue

	log *zap.Logger
}

func NewMemStorage(metricsMap map[string]server.StorageValue, log *zap.Logger) *MemStorage {
	return &MemStorage{
		metricsMap: metricsMap,
		log:        log,
	}
}

func (m *MemStorage) UpdateValue(metricName string, metricValue float64) {
	m.log.Info("Update in storage")
	m.metricsMap[metricName] = server.StorageValue{Gauge: metricValue}
	for k, v := range m.metricsMap {
		fmt.Println(k, v)
	}
}

func (m *MemStorage) SumValue(metricName string, metricValue int64) {
	if value, exists := m.metricsMap[metricName]; exists {
		m.metricsMap[metricName] = server.StorageValue{
			Counter: value.Counter + metricValue,
		}
	} else {
		m.metricsMap[metricName] = server.StorageValue{
			Counter: metricValue,
		}
	}
	for k, v := range m.metricsMap {
		fmt.Println(k, v)
	}
}
