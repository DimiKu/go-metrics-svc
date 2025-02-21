package storage

import (
	"go-metric-svc/dto"
	"go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"strconv"
)

type MemStorage struct {
	metricsMap map[string]models.StorageValue

	log *zap.SugaredLogger
}

func NewMemStorage(metricsMap map[string]models.StorageValue, log *zap.SugaredLogger) *MemStorage {
	return &MemStorage{
		metricsMap: metricsMap,
		log:        log,
	}
}

func (m *MemStorage) UpdateValue(metricName string, metricValue float64) {
	//m.log.Info("Update in storage")
	m.metricsMap[metricName] = models.StorageValue{Gauge: metricValue}
}

func (m *MemStorage) SumValue(metricName string, metricValue int64) {
	if value, exists := m.metricsMap[metricName]; exists {
		m.metricsMap[metricName] = models.StorageValue{
			Counter: value.Counter + metricValue,
		}
	} else {
		m.metricsMap[metricName] = models.StorageValue{
			Counter: metricValue,
		}
	}
}

func (m *MemStorage) GetMetricByName(metric dto.MetricServiceDto) (dto.MetricServiceDto, error) {
	storageMetric := dto.MetricStorageDto(metric)

	if _, exists := m.metricsMap[storageMetric.Name]; !exists {
		return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
	}

	if storageMetric.MetricType == dto.MetricTypeHandlerCounterTypeDto {
		storageMetric.Value = strconv.FormatInt(m.metricsMap[storageMetric.Name].Counter, 10)
	} else if storageMetric.MetricType == dto.MetricTypeHandlerGaugeTypeDto {
		storageMetric.Value = strconv.FormatFloat(m.metricsMap[storageMetric.Name].Gauge, 'f', -1, 64)
	}

	return dto.MetricServiceDto(storageMetric), nil

}

func (m *MemStorage) GetAllMetrics() []string {
	metricSlide := make([]string, len(m.metricsMap))
	for k := range m.metricsMap {
		metricSlide = append(metricSlide, k)
	}
	return metricSlide
}
