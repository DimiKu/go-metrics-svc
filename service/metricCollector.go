package service

import (
	"go.uber.org/zap"
)

type Storage interface {
	UpdateValue(metricName string, metricValue float64)
	SumValue(metricName string, metricValue int64)
}

type MetricCollectorSvc struct {
	memStorage Storage

	log *zap.Logger
}

func NewMetricCollectorSvc(
	memStorage Storage,
	log *zap.Logger,
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
