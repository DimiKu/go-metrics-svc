package server

import (
	"go-metric-svc/dto"
	"go-metric-svc/internal/customErrors"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateValue(metricName string, metricValue float64)
	SumValue(metricName string, metricValue int64)
	GetMetricByName(metricName dto.MetricServiceDto) (dto.MetricServiceDto, error)
	GetAllMetrics() []string
}

type MetricCollectorSvc struct {
	memStorage Storage

	log *zap.SugaredLogger
}

func NewMetricCollectorSvc(
	memStorage Storage,
	log *zap.SugaredLogger,
) *MetricCollectorSvc {
	return &MetricCollectorSvc{
		memStorage: memStorage,
		log:        log,
	}
}

func (s *MetricCollectorSvc) UpdateStorage(metricName string, metricValue float64) {
	s.log.Info("Update in service")

	s.memStorage.UpdateValue(metricName, metricValue)
}

func (s *MetricCollectorSvc) SumInStorage(metricName string, metricValue int64) {
	s.log.Info("Sum metric in service")
	s.memStorage.SumValue(metricName, metricValue)
}

func (s *MetricCollectorSvc) GetMetricByName(metric dto.MetricServiceDto) (dto.MetricServiceDto, error) {
	collectedMetric, err := s.memStorage.GetMetricByName(metric)
	if err != nil {
		return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
	}
	return collectedMetric, nil
}

func (s *MetricCollectorSvc) GetAllMetrics() []string {
	return s.memStorage.GetAllMetrics()
}
