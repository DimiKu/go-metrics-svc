package storage

import (
	"fmt"
	"go.uber.org/zap"
	"strconv"
)

type MemStorage struct {
	metricsMap map[string]string

	log *zap.Logger
}

func NewMemStorage(metricsMap map[string]string, log *zap.Logger) *MemStorage {
	return &MemStorage{
		metricsMap: metricsMap,
		log:        log,
	}
}

func (m *MemStorage) UpdateValue(metricName string, metricValue string) {
	m.log.Info("Update in storage")
	m.metricsMap[metricName] = metricValue
	fmt.Println(m.metricsMap)
}

func (m *MemStorage) SumValue(metricName string, metricValue string) {
	if _, exists := m.metricsMap[metricName]; exists {
		oldValue, err := strconv.Atoi(m.metricsMap[metricName])
		if err != nil {
			m.log.Error("Cant convert m.metricsMap[metricName]")
		}
		newValue, err := strconv.Atoi(metricValue)
		if err != nil {
			m.log.Error("Cant convert metricValue")
		}
		newMapValue := oldValue + newValue
		m.metricsMap[metricName] = fmt.Sprintf("%d", newMapValue)
	} else {
		m.metricsMap[metricName] = metricValue
	}
	fmt.Println(m.metricsMap)
}
