package storage

import (
	"context"
	"go-metric-svc/dto"
	"go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"strconv"
)

// MemStorage структура реализующая сбор метик и их обновление

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

func (m *MemStorage) UpdateValue(metricName string, metricValue float64, ctx context.Context) error {
	//m.log.Info("Update in storage")
	m.metricsMap[metricName] = models.StorageValue{Gauge: metricValue}
	m.log.Infof("storage is: %s", m.metricsMap)
	return nil
}

func (m *MemStorage) SumValue(metricName string, metricValue int64, ctx context.Context) (int64, error) {
	if value, exists := m.metricsMap[metricName]; exists {
		m.metricsMap[metricName] = models.StorageValue{
			Counter: value.Counter + metricValue,
		}
	} else {
		m.metricsMap[metricName] = models.StorageValue{
			Counter: metricValue,
		}
	}

	return m.metricsMap[metricName].Counter, nil
}

func (m *MemStorage) GetMetricByName(metric dto.MetricServiceDto, ctx context.Context) (dto.MetricServiceDto, error) {
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

func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]string, error) {
	metricSlide := make([]string, len(m.metricsMap))
	for k := range m.metricsMap {
		metricSlide = append(metricSlide, k)
	}
	return metricSlide, nil
}
func (m *MemStorage) DBPing(ctx context.Context) (bool, error) {
	return false, nil
}

func (m *MemStorage) SaveMetrics(ctx context.Context, metrics dto.MetricCollectionDto) error {
	for _, metric := range metrics.CounterCollection {
		value, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return err
		}
		if _, err = m.SumValue(metric.Name, value, ctx); err != nil {
			return err
		}
	}

	for _, metric := range metrics.GaugeCollection {
		value, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return err
		}

		if err := m.UpdateValue(metric.Name, value, ctx); err != nil {
			return err
		}
	}

	return nil
}
